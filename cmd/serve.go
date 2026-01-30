package cmd

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"yggdrasil/sim-words/internal/cluster"
	"yggdrasil/sim-words/internal/search"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB
var clusters = []cluster.Cluster{}

func RunServe(args []string) {
	serveCmd := flag.NewFlagSet("query", flag.ExitOnError)

	port := serveCmd.Int("p", 3000, "server port")
	dbFilePath := serveCmd.String("db", "data.sqlite", "path to storage data")

	serveCmd.Parse(args)

	var err error
	db, err = gorm.Open(sqlite.Open(*dbFilePath), &gorm.Config{})
	if err != nil {
		log.Fatalf("unable to open db connection: %s", err)
	}
	log.Println("database inited")

	// 读取簇
	clusters, err = cluster.GetClusters(db, nil)
	if err != nil {
		log.Fatalf("unable to get clusters: %s", err)
		return
	}
	log.Printf("read %d clusters", len(clusters))

	r := gin.Default()

	r.Use(ErrorHandler())

	r.GET("/query", handleQuery)

	r.Run(fmt.Sprintf(":%d", *port))
}

func handleQuery(c *gin.Context) {
	kStr := c.DefaultQuery("k", "3")
	lStr := c.DefaultQuery("l", "5")
	query := c.DefaultQuery("q", "")
	template := c.DefaultQuery("t", "")
	log.Printf("query k=%s l=%s q=%s t=%s", kStr, lStr, query, template)

	k, err := strconv.Atoi(kStr)
	if err != nil {
		c.Error(fmt.Errorf("%s is not a valid number", kStr))
	}

	l, err := strconv.Atoi(lStr)
	if err != nil {
		c.Error(fmt.Errorf("%s is not a valid number", lStr))
	}

	// 查询
	var results []search.SearchResult
	if template == "" {
		embd, err := embedWord(query)
		if err != nil {
			c.Error(fmt.Errorf("unable to embed query string: %s", err))
		}

		results, err = search.QueryWords(db, embd, clusters, k, l, false)
		if err != nil {
			c.Error(fmt.Errorf("unable to query words: %s", err))
			return
		}
	} else {
		log.Printf("query with template: %s", template)
		results, err = search.QueryWordsWithTemplate(db, query, template, clusters, k, l, false)
		if err != nil {
			c.Error(fmt.Errorf("unable to query words: %s", err))
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "ok",
		"data":    results,
	})
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			c.JSON(http.StatusInternalServerError, map[string]any{
				"success": false,
				"message": err.Error(),
			})
		}
	}
}
