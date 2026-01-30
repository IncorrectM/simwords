package search

import (
	"fmt"
	"math"
	"sort"
	"yggdrasil/sim-words/internal/base"
	"yggdrasil/sim-words/internal/cluster"
	"yggdrasil/sim-words/internal/word"

	"gorm.io/gorm"
)

type SearchResult struct {
	Word       string
	Similarity float64
	Frequency  int
}

func CosineSimilarity(a, b base.Float64Slice) float64 {
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// QueryWords 查询相似单词
func QueryWords(
	db *gorm.DB,
	query base.Float64Slice,
	clusters []cluster.Cluster, // 每个簇里包含 Center 和 Words
	topK int,
	L int,
	includeSelf bool,
) ([]SearchResult, error) {

	type clusterScore struct {
		ClusterIndex int
		Score        float64
	}

	// 计算 query 与簇中心相似度
	scores := make([]clusterScore, len(clusters))
	for i, c := range clusters {
		scores[i] = clusterScore{
			ClusterIndex: i,
			Score:        CosineSimilarity(query, c.Embedding.NormalizedEmbedding),
		}
	}

	//排序找 top-K（最相似）和 bottom-K（最远）
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})
	topClusters := scores[:topK]
	bottomClusters := scores[len(scores)-topK:] // 同样取 topK 个最远

	// 从 topClusters 中选相似度最高的 L 个单词
	var results []SearchResult
	const epsilon = 1e-6
	selectTopL := func(clist []clusterScore) error {
		for _, cs := range clist {
			c := clusters[cs.ClusterIndex]
			// 在簇内所有单词计算相似度
			words, err := word.SelectByClusterID(db, c.ID)
			if err != nil {
				return fmt.Errorf("unable to load words in cluster %d: %s", c.ID, err)
			}
			sort.Slice(words, func(i, j int) bool {
				return CosineSimilarity(query, words[i].NormalizedEmbedding) >
					CosineSimilarity(query, words[j].NormalizedEmbedding)
			})

			count := 0
			for i := 0; count < L && i < len(words); i++ {
				sim := CosineSimilarity(query, words[i].NormalizedEmbedding)

				if !includeSelf {
					// 不允许包含自己，则判断是不是自己
					if math.Abs(sim-1.0) < epsilon {
						// 当差值很小时，视作自己
						continue
					}
				}

				results = append(results, SearchResult{
					Word:       words[i].Word,
					Similarity: sim,
					Frequency:  words[i].Frequency,
				})
				count++
			}
		}
		return nil
	}

	err := selectTopL(topClusters)
	if err != nil {
		return nil, fmt.Errorf("unable to load from top clusters: %s", err)
	}
	err = selectTopL(bottomClusters)
	if err != nil {
		return nil, fmt.Errorf("unable to load from bottom clusters: %s", err)
	}

	return results, nil
}
