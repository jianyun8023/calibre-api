package main

import (
	"github.com/meilisearch/meilisearch-go"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

var client = meilisearch.NewClient(meilisearch.ClientConfig{
	Host:   "http://192.168.2.4:7700",
	APIKey: "",
})

// An index is where the documents are stored.
var books = client.Index("books")

var db = make(map[string]string)

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Get user value
	r.GET("/user/:name", func(c *gin.Context) {
		user := c.Params.ByName("name")
		value, ok := db[user]
		if ok {
			c.JSON(http.StatusOK, gin.H{"user": user, "value": value})
		} else {
			c.JSON(http.StatusOK, gin.H{"user": user, "status": "no value"})
		}
	})

	r.GET("/get/cover/:id", func(c *gin.Context) {
		id := strings.TrimSuffix(c.Param("id"), ".jpg")

		var book Book
		err := books.GetDocument(id, nil, &book)
		//var fs http.FileSystem = // ...

		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
		} else {
			//c.FileFromFS("fs/file.go", fs)
			c.Redirect(http.StatusFound, "http://192.168.2.4:8188/get/cover/"+id+".jpg")
		}
	})

	r.GET("/book/:id", func(c *gin.Context) {
		id := c.Param("id")
		var book Book
		err := books.GetDocument(id, nil, &book)

		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
		} else {
			c.JSON(http.StatusOK, book)
		}
	})

	r.GET("/search", func(c *gin.Context) {
		q := c.Query("q")
		limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)
		offset, _ := strconv.ParseInt(c.DefaultQuery("Offset", "0"), 10, 64)
		search, err := books.Search(q, &meilisearch.SearchRequest{
			Offset: offset,
			Limit:  limit,
			Sort:   []string{"id:desc"},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
		}

		c.JSON(http.StatusOK, &search)
	})

	// Authorized group (uses gin.BasicAuth() middleware)
	// Same than:
	// authorized := r.Group("/")
	// authorized.Use(gin.BasicAuth(gin.Credentials{
	//	  "foo":  "bar",
	//	  "manu": "123",
	//}))
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"foo":  "bar", // user:foo password:bar
		"manu": "123", // user:manu password:123
	}))

	/* example curl for /admin with basicauth header
	   Zm9vOmJhcg== is base64("foo:bar")

		curl -X POST \
	  	http://localhost:8080/admin \
	  	-H 'authorization: Basic Zm9vOmJhcg==' \
	  	-H 'content-type: application/json' \
	  	-d '{"value":"bar"}'
	*/
	authorized.POST("admin", func(c *gin.Context) {
		user := c.MustGet(gin.AuthUserKey).(string)

		// Parse JSON
		var json struct {
			Value string `json:"value" binding:"required"`
		}

		if c.Bind(&json) == nil {
			db[user] = json.Value
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		}
	})

	return r
}

func main() {
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}
