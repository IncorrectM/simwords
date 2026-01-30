package cmd

import (
	"flag"
	"log"
	"yggdrasil/sim-words/internal/base"
	"yggdrasil/sim-words/internal/cluster"
	"yggdrasil/sim-words/internal/embedding"
	"yggdrasil/sim-words/internal/search"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func RunQuery(args []string) {
	queryCmd := flag.NewFlagSet("query", flag.ExitOnError)

	k := queryCmd.Int("k", 3, "select top k clusters")
	l := queryCmd.Int("l", 5, "select top l words in the cluster")

	query := queryCmd.String("q", "", "keyword to query")

	dbFilePath := queryCmd.String("db", "data.sqlite", "path to storage data")

	queryCmd.Parse(args)

	// 强制非空检查
	if *query == "" {
		log.Fatalln("query cannot be empty. Use -q <keyword>")
	}
	log.Printf("query %s with k=%d, l=%d", *query, *k, *l)

	// 嵌入化查询字符
	embd, err := embedWord(*query)
	if err != nil {
		log.Fatalf("unable to embed query string: %s", *query)
	}

	// 初始化数据库连接
	db, err := gorm.Open(sqlite.Open(*dbFilePath), &gorm.Config{})
	if err != nil {
		log.Fatalf("unable to open db connection: %s", err.Error())
	}
	log.Printf("database inited")

	// 读取簇
	clusters, err := cluster.GetClusters(db, nil)
	if err != nil {
		log.Fatalf("unable to get clusters: %s", err)
	}
	log.Printf("read %d clusters", len(clusters))

	// 查询
	results, err := search.QueryWords(db, embd, clusters, *k, *l)
	if err != nil {
		log.Fatalf("unable to query words: %s", err)
	}
	for _, r := range results {
		log.Printf("%s\t%.2g", r.Word, r.Similarity)
	}
}

func embedWord(str string) (base.Float64Slice, error) {
	value, err := embedding.Embedding([]string{str})
	if err != nil {
		return nil, err
	}
	return value.Embeddings[0], nil
}
