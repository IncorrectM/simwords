package search

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"yggdrasil/sim-words/internal/base"
	"yggdrasil/sim-words/internal/cluster"
	"yggdrasil/sim-words/internal/common"
	"yggdrasil/sim-words/internal/embedding"
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
	clusters []cluster.Cluster,
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

	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})
	return results, nil
}

func QueryWordsWithTemplate(
	db *gorm.DB,
	query string,
	template string,
	clusters []cluster.Cluster,
	topK int,
	L int,
	includeSelf bool,
) ([]SearchResult, error) {
	type clusterScore struct {
		ClusterIndex int
		Score        float64
	}

	type templatedCluster struct {
		cluster.Cluster
		TemplatedEmbedding base.Float64Slice
	}

	clusterInputs := make([]string, len(clusters)+1)
	clusterInputs[0] = strings.ReplaceAll(template, "{{placeholder}}", query)
	for i, c := range clusters {
		clusterInputs[i+1] = strings.ReplaceAll(template, "{{placeholder}}", c.AnchorWord)
	}
	clusterEmbeddings, err := embedding.Embedding(clusterInputs)
	if err != nil {
		return nil, err
	}

	queryEmbedding := clusterEmbeddings.Embeddings[0]
	templatedClusters := make([]templatedCluster, len(clusters))
	for i, c := range clusters {
		templatedClusters[i] = templatedCluster{
			Cluster:            c,
			TemplatedEmbedding: clusterEmbeddings.Embeddings[i+1],
		}
	}

	// 计算 query 与簇中心相似度
	scores := make([]clusterScore, len(clusters))
	for i, c := range templatedClusters {
		scores[i] = clusterScore{
			ClusterIndex: i,
			Score:        CosineSimilarity(queryEmbedding, c.TemplatedEmbedding),
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
		for i, cs := range clist {
			c := templatedClusters[cs.ClusterIndex]
			log.Printf("cluster #%d anchor %s", i, c.AnchorWord)
			// 在簇内所有单词计算相似度
			words, err := word.SelectByClusterID(db, c.ID)
			if err != nil {
				return fmt.Errorf("unable to load words in cluster %d: %s", c.ID, err)
			}

			inputs := common.Map(words, func(w word.WordEmbedding) string {
				return strings.ReplaceAll(template, "{{placeholder}}", w.Word)
			})

			wordEmbeddings, err := embedding.Embedding(inputs)
			if err != nil {
				return err
			}

			count := 0
			for i := 0; count < L && i < len(words); i++ {
				sim := CosineSimilarity(queryEmbedding, wordEmbeddings.Embeddings[i])
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

	err = selectTopL(topClusters)
	if err != nil {
		return nil, fmt.Errorf("unable to load from top clusters: %s", err)
	}
	err = selectTopL(bottomClusters)
	if err != nil {
		return nil, fmt.Errorf("unable to load from bottom clusters: %s", err)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})
	return results, nil
}
