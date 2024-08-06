package calibre

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jianyun8023/calibre-api/pkg/client"
	"github.com/jianyun8023/calibre-api/pkg/content"
	"github.com/jianyun8023/calibre-api/pkg/log"
	"github.com/kapmahc/epub"
	"github.com/meilisearch/meilisearch-go"
	"github.com/spf13/cast"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

type Api struct {
	config     *Config
	contentApi *content.Api
	client     *meilisearch.Client
	bookIndex  *meilisearch.Index
	baseDir    string
	http       *client.Client
}

func (c Api) SetupRouter(r *gin.Engine) {

	base := r.Group("/api")
	base.GET("/get/cover/:id", c.getCover)
	base.GET("/proxy/cover/*path", c.proxyCover)
	base.GET("/get/book/:id", c.getBookFile)
	base.GET("/read/:id/toc", c.getBookToc)
	base.GET("/read/:id/file/*path", c.getBookContent)
	base.GET("/book/:id", c.getBook)
	base.POST("/book/:id/delete", c.deleteBook)
	base.POST("/book/:id/update", c.updateMetadata)
	base.GET("/search", c.search)
	base.GET("/metadata/isbn/:isbn", c.getIsbn)
	base.GET("/metadata/search", c.queryMetadata)
	base.POST("/search", c.search)
	// 最近更新Recently
	base.GET("/recently", c.recently)
	base.POST("/index/update", c.updateIndex)
}

func NewClient(config *Config) Api {
	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   config.Search.Host,
		APIKey: config.Search.APIKey,
	})

	baseDir := config.TmpDir
	if !Exists(baseDir) {
		os.MkdirAll(baseDir, fs.ModePerm)
	}

	//index := client.Index(config.Search.Index)
	index, err := ensureIndexExists(client, config.Search.Index)
	if err != nil {
		log.Fatal(err)
	}
	_, err = ensureIndexExists(client, config.Search.Index+"-bak")
	if err != nil {
		log.Fatal(err)
	}
	newClient, err := content.NewClient(config.Content.Server)
	if err != nil {
		log.Fatal(err)
	}
	return Api{
		config:     config,
		client:     client,
		bookIndex:  index,
		baseDir:    config.TmpDir,
		contentApi: &newClient,
		http:       newClient.Client,
	}
}

// ensureIndexExists checks if a Meilisearch index exists, and if not, creates it and updates its settings.
//
// Parameters:
// - client: A pointer to the Meilisearch client.
// - indexName: The name of the index to check or create.
//
// Returns:
// - A pointer to the Meilisearch index.
// - An error if the index creation or settings update fails.
func ensureIndexExists(client *meilisearch.Client, indexName string) (*meilisearch.Index, error) {
	index := client.Index(indexName)

	// Fetch index information to check if it exists
	log.Infof("Checking if index %q exists", indexName)
	_, err := index.FetchInfo()
	if err != nil {
		log.Infof("Failed to fetch index info for %q: %v", indexName, err)
		// Index does not exist, create it
		log.Infof("Creating index %q", indexName)
		_, err = client.CreateIndex(&meilisearch.IndexConfig{
			Uid:        indexName,
			PrimaryKey: "id",
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create index: %w", err)
		}
		log.Infof("Index %q created", indexName)
		// Update index settings
		log.Infof("Updating index settings for %q", indexName)
		_, err = index.UpdateSettings(&meilisearch.Settings{
			DisplayedAttributes:  []string{"*"},
			FilterableAttributes: []string{"authors", "file_path", "id", "last_modified", "pubdate", "publisher", "isbn", "tags"},
			SearchableAttributes: []string{"title", "authors", "isbn"},
			SortableAttributes:   []string{"authors_sort", "id", "last_modified", "pubdate", "publisher"},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update index settings: %w", err)
		}
	}
	return index, nil
}

func (c Api) search(r *gin.Context) {
	var req = meilisearch.SearchRequest{}
	err2 := r.Bind(&req)
	if err2 != nil {
		log.Infof("====== Only Bind By Query String ======\n%v", err2)
	}
	if len(req.Sort) == 0 {
		req.Sort = []string{"id:desc"}
	}
	log.Infof("search request: %v", req)
	q := r.Query("q")
	if q == "" {
		q = r.PostForm("q")
	}
	log.Infof("search query: %s", q)
	search, err := c.bookIndex.Search(q, &req)

	books := make([]Book, len(search.Hits))
	for i := range search.Hits {
		tmp := search.Hits[i].(map[string]interface{})
		jsonb, err := json.Marshal(tmp)
		if err != nil {
			// do error check
			fmt.Println(err)
			return
		}

		book := Book{}
		if err := json.Unmarshal(jsonb, &book); err != nil {
			// do error check
			fmt.Println(err)
			return
		}
		books[i] = book
	}

	if err != nil {
		r.JSON(http.StatusInternalServerError, err)
	}

	r.JSON(http.StatusOK, gin.H{
		"estimatedTotalHits": search.EstimatedTotalHits,
		"offset":             search.Offset,
		"limit":              search.Limit,
		"processingTimeMs":   search.ProcessingTimeMs,
		"query":              search.Query,
		"hits":               &books,
	})
}

func (c Api) getBook(r *gin.Context) {
	id := r.Param("id")
	var book Book
	err := c.bookIndex.GetDocument(id, nil, &book)

	if err != nil {
		// 返回文件找不到
		r.JSON(http.StatusNotFound, "book not found")
		return
	}
	r.JSON(http.StatusOK, book)

}

func (c Api) deleteBook(r *gin.Context) {
	id := r.Param("id")

	err := c.contentApi.DeleteBooks([]string{id}, "")
	if err != nil {
		r.JSON(http.StatusNotFound, "book not found"+err.Error())
		return
	}
	_, err = c.bookIndex.DeleteDocument(id)
	if err != nil {
		// 返回文件找不到
		r.JSON(http.StatusNotFound, "book not found"+err.Error())
		return
	}
	r.JSON(http.StatusOK, "success")
}

func (c Api) getBookToc(r *gin.Context) {
	id := strings.TrimSuffix(r.Param("id"), ".epub")

	filepath, _ := c.getFileOrCache(id)
	book, _ := epub.Open(filepath)
	points := c.expansionTree(book.Ncx.Points)
	var p []epub.NavPoint
	for i := range points {
		point := points[i]
		p = append(p, epub.NavPoint{
			Text: point.Text,
			Content: epub.Content{
				Src: path.Join("/read/"+id+"/file", path.Dir(book.Container.Rootfile.Path), point.Content.Src),
			},
		})
	}

	defer book.Close()

	r.JSON(http.StatusOK, gin.H{
		"points":   p,
		"metadata": book.Opf.Metadata,
		"manifest": book.Opf.Manifest,
		"baseDir":  path.Dir(book.Container.Rootfile.Path),
	})

}

func (c Api) expansionTree(ori []epub.NavPoint) []epub.NavPoint {
	var points []epub.NavPoint
	for i := range ori {
		point := ori[i]
		points = append(points, point)
		if len(point.Points) > 0 {
			points = append(points, c.expansionTree(point.Points)...)
		}
	}
	return points
}

func (c Api) getBookContent(r *gin.Context) {
	id := strings.TrimSuffix(r.Param("id"), ".epub")

	//path1 := path.Join(c.Query("baseDir"), c.Query("path"))
	path1 := r.Param("path")
	var book Book
	err := c.bookIndex.GetDocument(id, nil, &book)
	if err != nil {
		r.JSON(http.StatusInternalServerError, err)
	} else {
		filepath, _ := c.getFileOrCache(id)

		destDir := path.Join(c.baseDir, id)

		if Exists(destDir) {
			s, _ := ioutil.ReadDir(destDir)
			if len(s) == 0 {
				fmt.Println("empty")
			}
		} else {
			os.MkdirAll(destDir, fs.ModePerm)
		}

		err := unzipSource(filepath, destDir)
		if err != nil {
			r.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": err.Error(),
			})
		}
		r.FileFromFS(path1, http.Dir(destDir))
	}
}

func (c Api) getFile(id string) (int64, io.ReadCloser, error) {
	size, reader, err := c.contentApi.GetBook(id, "library")
	return size, reader, err
}

func (c Api) getBookFile(r *gin.Context) {
	filesuffix := path.Ext(r.Param("id"))
	id := strings.TrimSuffix(r.Param("id"), filesuffix)

	size, reader, err := c.contentApi.GetBook(id, "library")
	if err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}
	defer reader.Close()
	r.DataFromReader(http.StatusOK, size, "application/epub+zip", reader, nil)
}

func (c Api) getCover(r *gin.Context) {
	id := strings.TrimSuffix(r.Param("id"), ".jpg")
	size, reader, err := c.contentApi.GetCover(id, "library")
	if err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}
	defer reader.Close()
	r.DataFromReader(http.StatusOK, size, "image/jpeg", reader, nil)
}

func (c Api) getFileOrCache(id string) (string, error) {
	filename := path.Join(c.baseDir, id+".epub")
	_, err := os.Stat(filename)
	if Exists(filename) {
		return filename, nil
	}
	_, closer, err := c.getFile(id)
	if err != nil {
		return "", err
	}
	b, err := io.ReadAll(closer)
	if err != nil {
		return "", err
	}
	closer.Close()

	f, err := os.Create(filename)
	defer f.Close()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		_, err = f.Write(b)
	}
	return filename, err
}

func (c Api) updateIndex(c2 *gin.Context) {
	booksIds, err2 := c.contentApi.GetAllBooksIds()
	if err2 != nil {
		c2.JSON(http.StatusInternalServerError, err2)
		return
	}

	index := c.client.Index(c.config.Search.Index + "-bak")
	_, err := index.DeleteAllDocuments()
	if err != nil {
		log.Warn(err)
		c2.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}

	// 按 2000 分段 booksIds,查询书籍，更新索引
	var books []Book
	for i := 0; i < len(booksIds); i += 2000 {
		ids := booksIds[i:min(i+2000, len(booksIds))]
		log.Infof("update index %d [%d - %d]", i, ids[0], ids[len(ids)-1])

		data, err := c.contentApi.GetBookMetaDatas(ids, "")
		if err != nil {
			log.Warnf("get book metadata error: %v", err)
			c2.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
			return
		}
		books, err = convertContentBooks(data)
		if err != nil {
			c2.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
			return
		}
		_, err = index.UpdateDocumentsInBatches(books, len(ids))
		if err != nil {
			c2.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
			return
		}
	}

	_, _ = c.client.SwapIndexes(
		[]meilisearch.SwapIndexesParams{
			{
				Indexes: []string{c.config.Search.Index, c.config.Search.Index + "-bak"},
			},
		},
	)
	c2.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    len(booksIds),
	})
}

func (c Api) recently(r *gin.Context) {
	limit, err := strconv.Atoi(r.DefaultQuery("limit", "10"))
	if err != nil {
		r.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit"})
		return
	}
	offset, err := strconv.Atoi(r.DefaultQuery("offset", "0"))
	if err != nil {
		r.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset"})
		return
	}

	searchRequest := meilisearch.SearchRequest{
		Sort:   []string{"id:desc"},
		Limit:  int64(limit),
		Offset: int64(offset),
	}

	search, err := c.bookIndex.Search("", &searchRequest)
	if err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	books := make([]Book, len(search.Hits))
	for i := range search.Hits {
		tmp := search.Hits[i].(map[string]interface{})
		jsonb, err := json.Marshal(tmp)
		if err != nil {
			r.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		book := Book{}
		if err := json.Unmarshal(jsonb, &book); err != nil {
			r.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		books[i] = book
	}

	r.JSON(http.StatusOK, gin.H{
		"totalHits":          search.TotalHits,
		"totalPages":         search.TotalPages,
		"hitsPerPage":        search.HitsPerPage,
		"estimatedTotalHits": search.EstimatedTotalHits,
		"offset":             search.Offset,
		"limit":              search.Limit,
		"processingTimeMs":   search.ProcessingTimeMs,
		"query":              search.Query,
		"hits":               &books,
	})
}

func (c Api) updateMetadata(r *gin.Context) {
	id := r.Param("id")
	book := &Book{}
	err := r.Bind(book)
	if err != nil {
		r.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	oldBook := &Book{}
	err = c.bookIndex.GetDocument(id, nil, oldBook)
	if err != nil {
		r.JSON(http.StatusNotFound, gin.H{
			"code":  500,
			"error": err.Error(),
			"msg":   "元数据更新失败",
		})
		return
	}
	_, err = c.contentApi.UpdateMetaData(id, parseParams(book, oldBook), "")
	if err != nil {
		r.JSON(http.StatusNotFound, gin.H{
			"code":  500,
			"error": err.Error(),
			"msg":   "元数据更新失败",
		})
		return
	}

	data, err := c.contentApi.GetBookMetaDatas([]int64{cast.ToInt64(id)}, "")
	if err != nil {
		log.Warnf("get book metadata error: %v", err)
		r.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}
	books, err := convertContentBooks(data)
	if err != nil {
		r.JSON(http.StatusNotFound, gin.H{
			"code":  500,
			"error": err.Error(),
			"msg":   "书籍元数据翻译失败，请刷新索引",
		})
		return
	}
	_, err = c.bookIndex.UpdateDocuments(books)
	if err != nil {
		// 返回文件找不到
		r.JSON(http.StatusNotFound, gin.H{
			"code":  500,
			"error": err.Error(),
			"msg":   "元数据更新成功，但是索引更新失败，请刷新索引",
		})
		return
	}
	r.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
	return
}

func convertContentBooks(content []content.Book) ([]Book, error) {
	var books []Book
	for _, c := range content {
		book := Book{
			// Map fields from Content to Book
			AuthorSort:   c.AuthorSort,
			Authors:      c.Authors,
			Comments:     c.Comments,
			ID:           c.ID,
			Isbn:         c.Isbn,
			Languages:    c.Languages,
			LastModified: c.LastModified,
			PubDate:      c.PubDate,
			Publisher:    c.Publisher,
			SeriesIndex:  c.SeriesIndex,
			Size:         c.Size,
			Title:        c.Title,
			Tags:         c.Tags,
			Rating:       c.Rating,
			Identifiers:  c.Identifiers,
			Cover:        "/api/get/cover/" + strconv.FormatInt(c.ID, 10) + ".jpg",
			FilePath:     "/api/get/book/" + strconv.FormatInt(c.ID, 10) + ".epub",
		}
		books = append(books, book)
	}
	return books, nil
}

func convertContentToBooks(content map[string]content.Content) ([]Book, error) {
	var books []Book
	for id, c := range content {
		i, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return nil, err
		}

		book := Book{
			// Map fields from Content to Book
			AuthorSort:   c.AuthorSort,
			Authors:      c.Authors,
			Comments:     c.Comments,
			ID:           i,
			Isbn:         c.Isbn,
			Languages:    c.Languages,
			LastModified: c.LastModified,
			PubDate:      c.PubDate,
			Publisher:    c.Publisher,
			SeriesIndex:  c.SeriesIndex,
			Size:         c.Size,
			Title:        c.Title,
			Tags:         c.Tags,
			Rating:       c.Rating,
			Identifiers:  c.Identifiers,
			Cover:        "/api/get/cover/" + strconv.FormatInt(i, 10) + ".jpg",
			FilePath:     "/api/get/book/" + strconv.FormatInt(i, 10) + ".epub",
		}
		books = append(books, book)
	}
	return books, nil
}

func (c Api) getIsbn(c2 *gin.Context) {
	isbn := c2.Param("isbn")
	//https://douban_isbn.yihy8023.workers.dev/v2/book/isbn/9787308242936
	resp, err := http.Get("https://douban_isbn.yihy8023.workers.dev/v2/book/isbn/" + isbn)
	if err != nil {
		c2.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		c2.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c2.JSON(http.StatusOK, body)
}

func (c Api) queryMetadata(c2 *gin.Context) {
	query := c2.Query("query")
	//url encode
	var jsonData map[string]interface{}
	resp, err := c.http.R().SetResult(&jsonData).SetQueryParam("q", query).Get(c.config.Metadata.DoubanUrl + "/v2/book/search")
	log.Infof(resp.Request.URL)
	if err != nil {
		c2.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c2.JSON(http.StatusOK, resp.Result())
}

func (c Api) proxyCover(r *gin.Context) {
	path := strings.TrimPrefix(r.Param("path"), "/")
	log.Infof("proxy cover: %s", path)
	response, err := c.http.R().SetDoNotParseResponse(true).
		SetHeader("Content-Type", "image/jpeg").
		SetQueryParamsFromValues(r.Request.URL.Query()).
		Get(path)
	if err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp := response.RawResponse
	length := resp.ContentLength
	reader := resp.Body
	defer reader.Close()
	r.DataFromReader(http.StatusOK, length, "image/jpeg", reader, nil)
}

func parseParams(book *Book, oldBook *Book) map[string]interface{} {
	///cdb/delete-books/264728/library
	metadata := map[string]interface{}{}
	if book.Comments != "" {
		metadata["comments"] = book.Comments
	}
	if book.Isbn != "" {
		identifiers := oldBook.Identifiers
		identifiers["isbn"] = book.Isbn
		metadata["identifiers"] = identifiers
	}
	if book.Title != "" {
		metadata["title"] = book.Title
	}
	if book.Publisher != "" {
		metadata["publisher"] = book.Publisher
	}
	//pubdate:"2024-05-01T12:00:00+00:00"

	if !book.PubDate.IsZero() {
		metadata["pubdate"] = book.PubDate.Format("2006-01-02T15:04:05+00:00")
	}
	if book.Authors != nil {
		metadata["authors"] = book.Authors
	}
	if book.Tags != nil {
		metadata["tags"] = book.Tags
	}
	if book.Rating > 0 {
		metadata["rating"] = book.Rating
	}
	return metadata
}
