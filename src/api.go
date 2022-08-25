package calibreApi

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kapmahc/epub"
	"github.com/meilisearch/meilisearch-go"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

type CalibreApi struct {
	config     *Config
	client     *meilisearch.Client
	bookIndex  *meilisearch.Index
	fileClient FileClient
	baseDir    string
}

func NewClient(config *Config) CalibreApi {
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
	return CalibreApi{
		config:     config,
		client:     client,
		bookIndex:  client.Index(config.Search.Index),
		fileClient: fileClient,
		baseDir:    config.Storage.TmpDir,
	}
}

func (api CalibreApi) Search(c *gin.Context) {
	var req = meilisearch.SearchRequest{}
	err2 := c.Bind(&req)
	if err2 != nil {
		log.Println("====== Only Bind By Query String ======", err2)
	}
	if len(req.Sort) == 0 {
		req.Sort = []string{"id:desc"}
	}
	q := c.Query("q")
	search, err := api.bookIndex.Search(q, &req)

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
		book.Cover = "/get/cover/" + strconv.FormatInt(id, 10) + ".jpg"
		for i := range book.Formats {
			s := book.Formats[i]
			book.Formats[i] = "/get/book/" + strconv.FormatInt(id, 10) + path.Ext(s)
		}
		books[i] = book
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}

	c.JSON(http.StatusOK, gin.H{
		"estimatedTotalHits": search.EstimatedTotalHits,
		"offset":             search.Offset,
		"limit":              search.Limit,
		"processingTimeMs":   search.ProcessingTimeMs,
		"query":              search.Query,
		"hits":               &books,
	})
}

func (api CalibreApi) GetBook(c *gin.Context) {
	id := c.Param("id")
	var book Book
	err := api.bookIndex.GetDocument(id, nil, &book)

	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}
	book.Cover = "/get/cover/" + id + ".jpg"
	for i := range book.Formats {
		s := book.Formats[i]
		book.Formats[i] = "/get/book/" + id + path.Ext(s)
	}
	c.JSON(http.StatusOK, book)

}

func (api CalibreApi) GetBookToc(c *gin.Context) {
	id := strings.TrimSuffix(c.Param("id"), ".epub")
	var book Book
	err := api.bookIndex.GetDocument(id, nil, &book)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	} else {
		filepath, _ := api.getFileOrCache(api.fixPath(book.Formats[0]), id)
		book, _ := epub.Open(filepath)
		points := api.expansionTree(book.Ncx.Points)
		var r []epub.NavPoint
		for i := range points {
			point := points[i]
			r = append(r, epub.NavPoint{
				Text: point.Text,
				Content: epub.Content{
					Src: path.Join("/read/"+id+"/file", path.Dir(book.Container.Rootfile.Path), point.Content.Src),
				},
			})
		}

		defer book.Close()
		c.JSON(http.StatusOK, gin.H{
			"points":   r,
			"metadata": book.Opf.Metadata,
			"manifest": book.Opf.Manifest,
			"baseDir":  path.Dir(book.Container.Rootfile.Path),
		})
	}
}

func (api CalibreApi) fixPath(s string) string {
	return strings.TrimPrefix(s, api.config.Search.TrimPath)
}

func (api CalibreApi) expansionTree(ori []epub.NavPoint) []epub.NavPoint {
	var points []epub.NavPoint
	for i := range ori {
		point := ori[i]
		points = append(points, point)
		if len(point.Points) > 0 {
			points = append(points, api.expansionTree(point.Points)...)
		}
	}
	return points
}

func (api CalibreApi) GetBookContent(c *gin.Context) {
	id := strings.TrimSuffix(c.Param("id"), ".epub")

	//path1 := path.Join(c.Query("baseDir"), c.Query("path"))
	path1 := c.Param("path")
	var book Book
	err := api.bookIndex.GetDocument(id, nil, &book)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	} else {
		p := api.fixPath(book.Formats[0])
		filepath, _ := api.getFileOrCache(p, id)

		destDir := path.Join(api.baseDir, id)

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
			c.JSON(http.StatusInternalServerError, err)
		}
		c.FileFromFS(path1, http.Dir(destDir))
	}
}

func (api CalibreApi) getFile(filepath string) (os.FileInfo, io.ReadCloser, error) {
	targetPath := path.Join(api.config.Storage.Webdav.Path, filepath)
	info, err := api.fileClient.Stat(targetPath)
	if err != nil {
		return nil, nil, err
	}
	reader, err := api.fileClient.ReadStream(targetPath)
	if err != nil {
		return nil, nil, err
	}
	return info, reader, nil
}

func (api CalibreApi) GetBookFile(c *gin.Context) {
	filesuffix := path.Ext(c.Param("id"))
	id := strings.TrimSuffix(c.Param("id"), filesuffix)
	var book Book
	err := api.bookIndex.GetDocument(id, nil, &book)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	} else {
		path := api.fixPath(book.Formats[0])
		info, reader, _ := api.getFile(path)
		defer reader.Close()
		c.DataFromReader(http.StatusOK, info.Size(), "", reader, nil)
	}
}

func (api CalibreApi) GetCover(c *gin.Context) {
	id := strings.TrimSuffix(c.Param("id"), ".jpg")

	var book Book
	err := api.bookIndex.GetDocument(id, nil, &book)

	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	} else {
		info, reader, _ := api.getFile(api.fixPath(book.Cover))
		defer reader.Close()
		c.DataFromReader(http.StatusOK, info.Size(), "", reader, nil)
	}
}

func (api CalibreApi) getFileOrCache(filepath string, id string) (string, error) {
	filename := path.Join(api.baseDir, id+".epub")
	_, err := os.Stat(filename)
	if Exists(filename) {
		return filename, nil
	}
	_, closer, err := api.getFile(filepath)
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
