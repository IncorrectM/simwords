package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode"
	"yggdrasil/sim-words/internal/base"
	"yggdrasil/sim-words/internal/cluster"
	"yggdrasil/sim-words/internal/common"
	"yggdrasil/sim-words/internal/embedding"
	"yggdrasil/sim-words/internal/kmeans"
	"yggdrasil/sim-words/internal/word"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func RunLoad(args []string) {
	loadCmd := flag.NewFlagSet("load", flag.ExitOnError)

	// flags
	inputPath := loadCmd.String("input", "", "input file path")
	loadCmd.StringVar(inputPath, "i", "", "shorthand for -input")

	minIndex := loadCmd.Int(
		"min-index",
		100,
		"skip records with index less than this value",
	)
	loadCmd.IntVar(minIndex, "mi", 100, "shorthand for -min-index")

	minFrequency := loadCmd.Int(
		"min-frequency",
		10,
		"keep records with frequency greater than this value",
	)
	loadCmd.IntVar(minFrequency, "mf", 10, "shorthand for -min-frequency")

	minLength := loadCmd.Int(
		"min-length",
		0,
		"keep records with word longer than this value",
	)
	loadCmd.IntVar(minLength, "ml", 0, "shorthand for min-length")

	// k-means flags
	k := loadCmd.Int("k", 10, "k of k-means")
	kIters := loadCmd.Int("kIters", 1000, "max iterations of k-means")

	// database flags
	dbFilePath := loadCmd.String("db", "data.sqlite", "path to storage data")

	// parse flags
	loadCmd.Parse(args)

	// initialized db connection
	db, err := gorm.Open(sqlite.Open(*dbFilePath), &gorm.Config{})
	if err != nil {
		log.Fatalf("unable to open db connection: %s", err.Error())
	}
	log.Printf("database inited")

	// load from given file
	rawRecords, err := loadFromFile(*inputPath, *minIndex, *minFrequency, *minLength)
	if err != nil {
		log.Fatalf("unable to load from %s: %s", *inputPath, err.Error())
	}
	log.Printf("read %d records", len(rawRecords))

	// embed words
	words, nil := embedWords(rawRecords, 1000)
	if err != nil {
		log.Fatalf("unable to embed: %s", err)
	}
	log.Printf("embed %d records", len(words))

	// nomalize embeddings
	for i := range words {
		words[i].NormalizedEmbedding = word.L2Normalize(words[i].RawEmbedding)
	}

	// save words
	err = word.SaveWords(db, words)
	if err != nil {
		log.Fatalf("unable to save words: %s", err)
	}
	log.Printf("saved %d words", len(words))

	// k-means clustering
	centers, clusterIndexies := kmeans.KMeans(words, *k, *kIters)
	clusters := common.Map(centers, func(vector []float64) cluster.Cluster {
		return cluster.Cluster{
			Embedding: base.Embedding{
				NormalizedEmbedding: vector,
			},
		}
	})

	// save clusters
	err = cluster.SaveClusters(db, clusters)
	if err != nil {
		log.Fatalf("unable to save clusters: %s", err)
	}
	log.Printf("saved %d clusters", len(clusters))

	// assign clusters to words
	err = word.BatchUpdateClusterIDs(db, words, common.Map(clusterIndexies, func(index uint) uint {
		return clusters[index].ID
	}))
	if err != nil {
		log.Fatalf("unable to update cluster IDs: %s", err)
	}
	log.Printf("%d words updated", len(words))
}

func loadFromFile(path string, minIndex int, minFrequency int, minLength int) ([]word.RawRecord, error) {
	freqMap := make(map[string]int)
	indexMap := make(map[string]int)

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			return nil, fmt.Errorf("invalid line: %s", line)
		}

		index, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, err
		}
		wordStr := cleanWord(strings.ToLower(parts[1]))
		frequency, err := strconv.Atoi(parts[2])
		if index <= minIndex || len(wordStr) < minLength {
			continue
		}

		freqMap[wordStr] += frequency
		if _, ok := indexMap[wordStr]; !ok {
			// 保留第一次出现的下标
			indexMap[wordStr] = index
		}
	}

	records := make([]word.RawRecord, 0, len(freqMap))
	for w, f := range freqMap {
		if f <= minFrequency {
			continue
		}

		index, ok := indexMap[w]
		if !ok {
			return nil, fmt.Errorf("cannot find index for word %s", w)
		}

		records = append(records, word.RawRecord{Index: index, Word: w, Frequency: f})
	}

	return records, nil
}

func embedWords(words []word.RawRecord, batchSize int) ([]word.WordEmbedding, error) {
	embeddings := make([]word.WordEmbedding, len(words))
	// 收集单词便于批量嵌入化
	wordStrs := make([]string, len(words))
	for i, w := range words {
		wordStrs[i] = w.Word
	}

	// 分批嵌入化
	batches := (len(words) + batchSize - 1) / batchSize // ceil(len/size)
	log.Printf("got %d batches", batches)

	for b := range batches {
		start := b * batchSize
		end := min((b+1)*batchSize, len(words))
		log.Printf("batch #%d: %d - %d", b, start, end)

		batch := words[start:end]
		batchEmbd, err := embedding.Embedding(common.Map(batch, func(record word.RawRecord) string {
			return record.Word
		}))
		if err != nil {
			return nil, err
		}
		for i, embd := range batchEmbd.Embeddings {
			index := b*batchSize + i
			embeddings[index] = word.WordEmbedding{
				Word:      words[index].Word,
				Frequency: words[index].Frequency,
				Embedding: base.Embedding{
					RawEmbedding: embd,
				},
			}
		}
	}

	return embeddings, nil
}

func cleanWord(raw string) string {
	var noSymbol strings.Builder
	for _, r := range raw {
		if unicode.IsLetter(r) {
			noSymbol.WriteRune(r)
		}
	}
	return noSymbol.String()
}
