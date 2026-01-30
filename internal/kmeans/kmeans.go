package kmeans

import (
	"log"
	"math"
	"math/rand"
	"runtime"
	"yggdrasil/sim-words/internal/base"
	"yggdrasil/sim-words/internal/word"
)

func KMeans(words []word.WordEmbedding, k int, maxIter int) ([][]float64, []uint) {
	n := len(words)
	dim := len(words[0].NormalizedEmbedding)

	// 1. 初始化中心
	centers := initCenters(words, k)

	clusterIDs := make([]uint, n)

	for itr := range maxIter {
		log.Printf("k-means iteration %d / %d", itr, maxIter)

		// 2. 分配阶段
		changed := assignClusters(words, centers, clusterIDs)

		// 3. 更新中心
		updateCenters(words, centers, clusterIDs, dim)

		if !changed {
			log.Println("k-means reached convergence")
			break
		}
	}

	return centers, clusterIDs
}

func initCenters(words []word.WordEmbedding, k int) [][]float64 {
	n := len(words)
	centers := make([][]float64, k)

	for i := range k {
		idx := rand.Intn(n)
		vec := make([]float64, len(words[idx].NormalizedEmbedding))
		copy(vec, words[idx].NormalizedEmbedding)
		centers[i] = vec
	}

	return centers
}

func assignClusters(words []word.WordEmbedding, centers [][]float64, clusterIDs []uint) bool {
	n := len(words)
	changed := false
	numWorkers := runtime.NumCPU()
	chunkSize := (n + numWorkers - 1) / numWorkers
	ch := make(chan bool, numWorkers)

	for w := range numWorkers {
		start := w * chunkSize
		end := min(start+chunkSize, n)
		log.Printf("launch chunk %d - %d", start, end)

		go func(start, end int) {
			localChanged := false
			for i := start; i < end; i++ {
				minDist := math.MaxFloat64
				var bestCluster uint = 0
				for j, center := range centers {
					dist := base.Distance(words[i].NormalizedEmbedding, center)
					if dist < minDist {
						minDist = dist
						bestCluster = uint(j)
					}
				}
				if clusterIDs[i] != bestCluster {
					clusterIDs[i] = bestCluster
					localChanged = true
				}
			}
			ch <- localChanged
		}(start, end)
	}

	for range numWorkers {
		if <-ch {
			changed = true
		}
	}

	return changed
}

func updateCenters(words []word.WordEmbedding, centers [][]float64, clusterIDs []uint, dim int) {
	k := len(centers)

	// 清零
	counts := make([]int, k)
	newCenters := make([][]float64, k)

	for i := range k {
		newCenters[i] = make([]float64, dim)
	}

	// 累加
	for i, word := range words {
		cluster := clusterIDs[i]
		counts[cluster]++

		for d := range dim {
			newCenters[cluster][d] += word.NormalizedEmbedding[d]
		}
	}

	// 求均值
	for i := range k {
		if counts[i] == 0 {
			continue
		}
		for d := range dim {
			newCenters[i][d] /= float64(counts[i])
		}
	}

	// 替换
	for i := range k {
		copy(centers[i], newCenters[i])
	}
}
