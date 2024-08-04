package calibre

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jianyun8023/calibre-api/pkg/client"
	"github.com/jianyun8023/calibre-api/pkg/content"
	"github.com/jianyun8023/calibre-api/pkg/log"
	"github.com/kapmahc/epub"
	"github.com/meilisearch/meilisearch-go"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

type Api struct {
	config     *Config
	contentApi *content.Api
	client     *meilisearch.Client
	bookIndex  *meilisearch.Index
	fileClient FileClient
	baseDir    string
}

func (c Api) SetupRouter(r *gin.Engine) {

	base := r.Group("/api")
	base.GET("/get/cover/:id", c.getCover)
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

	baseDir := config.Storage.TmpDir
	if !Exists(baseDir) {
		os.MkdirAll(baseDir, fs.ModePerm)
	}
	var fileClient FileClient
	switch config.Storage.Use {
	case "webdav":
		fileClient = NewWebDavClient(config.Storage.Webdav)
	case "local":
		fileClient = NewLocalClient(config.Storage.Local)
	case "minio":
		t, err := NewMinioClient(config.Storage.Minio, context.Background())
		fileClient = t
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal(fmt.Errorf("不支持的存储类型 %q", config.Storage.Use))
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
		fileClient: fileClient,
		baseDir:    config.Storage.TmpDir,
		contentApi: &newClient,
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
		id := book.ID
		book.Cover = "/api/get/cover/" + strconv.FormatInt(id, 10) + ".jpg"
		book.FilePath = "/api/get/book/" + strconv.FormatInt(id, 10) + ".epub"
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
	book.Cover = "/api/get/cover/" + id + ".jpg"
	book.FilePath = "/api/get/book/" + id + ".epub"
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
	var book Book
	err := c.bookIndex.GetDocument(id, nil, &book)
	if err != nil {
		r.JSON(http.StatusInternalServerError, err)
	} else {
		filepath, _ := c.getFileOrCache(book.FilePath, id)
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
}

func (c Api) fixPath(s string) string {
	return strings.TrimPrefix(s, c.config.Search.TrimPath)
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
		filepath, _ := c.getFileOrCache(book.FilePath, id)

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

func (c Api) getFile(filepath string) (os.FileInfo, io.ReadCloser, error) {
	targetPath := path.Join(c.config.Storage.Webdav.Path, filepath)
	info, err := c.fileClient.Stat(targetPath)
	if err != nil {
		return nil, nil, err
	}
	reader, err := c.fileClient.ReadStream(targetPath)
	if err != nil {
		return nil, nil, err
	}
	return info, reader, nil
}

func (c Api) stat(filepath string) (os.FileInfo, error) {
	targetPath := path.Join(c.config.Storage.Webdav.Path, filepath)
	info, err := c.fileClient.Stat(targetPath)
	return info, err
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

func (c Api) getFileOrCache(filepath string, id string) (string, error) {
	filename := path.Join(c.baseDir, id+".epub")
	_, err := os.Stat(filename)
	if Exists(filename) {
		return filename, nil
	}
	_, closer, err := c.getFile(filepath)
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadAll(closer)
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

func (c Api) getDbFileOrCache(flush bool) (string, error) {
	filename := path.Join(c.baseDir, "metadata.db")
	info, err := os.Stat(filename)
	remoteInfo, err := c.stat("metadata.db")
	if err != nil {
		return "", err
	}

	if !flush || Exists(filename) {
		log.Info("cached metadata.db")
		log.Infof("local file size: %d, remote file size: %d", info.Size(), remoteInfo.Size())
		log.Infof("local file time: %d, remote file time: %d", info.ModTime().Unix(), remoteInfo.ModTime().Unix())
		if info.ModTime().Unix() == remoteInfo.ModTime().Unix() && info.Size() == remoteInfo.Size() {
			return filename, nil
		} else {
			log.Info("remove cached metadata.db")
			_ = os.Remove(filename)
		}
	}
	_, closer, err := c.getFile("metadata.db")
	if err != nil {
		return "", err
	}
	b, err := io.ReadAll(closer)
	if err != nil {
		return "", err
	}
	defer closer.Close()

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
	dbPath, err := c.getDbFileOrCache(true)
	newDb, _ := NewDb(dbPath)
	books, _ := newDb.queryAllBooks()

	index := c.client.Index(c.config.Search.Index + "-bak")
	_, err = index.DeleteAllDocuments()
	if err != nil {
		log.Warn(err)
		c2.JSON(http.StatusInternalServerError, err)
		return
	}

	_, err = index.UpdateDocumentsInBatches(*books, 1000)
	if err != nil {
		log.Warn(err)
		c2.JSON(http.StatusInternalServerError, err)
		return
	}

	_, err = c.client.SwapIndexes(
		[]meilisearch.SwapIndexesParams{
			{
				Indexes: []string{c.config.Search.Index, c.config.Search.Index + "-bak"},
			},
		},
	)
	c2.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    len(*books),
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
		id := book.ID
		book.Cover = "/api/get/cover/" + strconv.FormatInt(id, 10) + ".jpg"
		book.FilePath = "/api/get/book/" + strconv.FormatInt(id, 10) + ".epub"
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

	data, err := c.contentApi.UpdateMetaData(id, parseParams(book), "")
	if err != nil {
		r.JSON(http.StatusNotFound, gin.H{
			"code":  500,
			"error": err.Error(),
			"msg":   "元数据更新失败",
		})
		return
	}

	var books []Book
	books, err = convertContentToBooks(data)
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
}

func convertContentToBooks(content map[string]content.Content) ([]Book, error) {
	var books []Book
	for id, c := range content {
		i, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return nil, err
		}

		// 2023-12-31T16:00:00+00:00 to time.Time
		//c.LastModified
		//c.Pubdate

		lastModified, err := time.Parse(time.RFC3339, c.LastModified)
		if err != nil {
			return nil, err
		}

		pubdate, err := time.Parse(time.RFC3339, c.Pubdate)
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
			LastModified: lastModified,
			Pubdate:      pubdate,
			Publisher:    c.Publisher,
			SeriesIndex:  c.SeriesIndex,
			Size:         c.Size,
			Title:        c.Title,
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
	//http://192.168.2.236:8085/v2/book/search?q=xxx
	//url encode

	api, err := client.New(&client.Config{
		Host:  "192.168.2.236:8085",
		HTTPS: false,
	})
	if err != nil {
		c2.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var jsonData map[string]interface{}

	resp, err := api.R().SetResult(&jsonData).SetQueryParam("q", query).Get("/v2/book/search")
	log.Infof(resp.Request.URL)
	if err != nil {
		c2.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c2.JSON(http.StatusOK, resp.Result())
}

func parseParams(book *Book) map[string]interface{} {
	///cdb/delete-books/264728/library
	metadata := map[string]interface{}{}
	if book.Comments != "" {
		metadata["comments"] = book.Comments
	}
	if book.Isbn != "" {
		metadata["identifiers"] = map[string]string{"isbn": book.Isbn}
	}
	if book.Title != "" {
		metadata["title"] = book.Title
	}
	if book.Publisher != "" {
		metadata["publisher"] = book.Publisher
	}
	//pubdate:"2024-05-01T12:00:00+00:00"

	if !book.Pubdate.IsZero() {
		metadata["pubdate"] = book.Pubdate.Format("2006-01-02T15:04:05+00:00")
	}
	if book.Authors != nil {
		metadata["authors"] = book.Authors
	}
	return metadata
}
