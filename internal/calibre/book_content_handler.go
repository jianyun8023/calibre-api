package calibre

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jianyun8023/calibre-api/internal/semantic/qdrant"
	"github.com/kapmahc/epub"
)

// getCover 获取书籍封面
func (c *Api) getCover(r *gin.Context) {
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

// proxyCover 代理封面图片
func (c *Api) proxyCover(r *gin.Context) {
	path := strings.TrimPrefix(r.Param("path"), "/")
	response, err := c.http.R().SetDoNotParseResponse(true).
		SetHeader("Referer", "https://book.douban.com/").
		SetHeader("Content-Type", "image/jpeg").
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3573.0 Safari/537.36").
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

// getBookFile 下载书籍文件
func (c *Api) getBookFile(r *gin.Context) {
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

// getBookToc 获取书籍目录
// 优先从 Qdrant 获取 TOC，缺失时从 EPUB 文件提取并自动更新到 Qdrant
func (c *Api) getBookToc(r *gin.Context) {
	id := strings.TrimSuffix(r.Param("id"), ".epub")
	toc, err := c.GetBookTocData(id)
	if err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}
	r.JSON(http.StatusOK, toc)
}

// GetBookTocData 获取书籍目录数据
func (c *Api) GetBookTocData(id string) (interface{}, error) {
	// Try to get TOC from Qdrant first
	if c.semanticSearcher != nil {
		if searcher, ok := c.semanticSearcher.(*qdrant.Searcher); ok {
			bookID := stringToInt64(id)
			if bookID > 0 {
				toc, err := searcher.GetBookToc(bookID)
				if err == nil && toc != nil {
					// TOC found in Qdrant, return it
					return toc, nil
				}
			}
		}
	}

	// TOC not found in Qdrant, extract from EPUB file
	tocData, err := c.extractTocFromEpub(id)
	if err != nil {
		return nil, fmt.Errorf("failed to extract TOC: %v", err)
	}

	// Asynchronously update TOC to Qdrant (don't block response)
	if c.semanticSearcher != nil {
		if searcher, ok := c.semanticSearcher.(*qdrant.Searcher); ok {
			bookID := stringToInt64(id)
			if bookID > 0 {
				go func() {
					if err := searcher.UpdateToc(bookID, tocData); err != nil {
						fmt.Printf("Warning: failed to update TOC in Qdrant for book %s: %v\n", id, err)
					}
				}()
			}
		}
	}

	return tocData, nil
}

// extractTocFromEpub extracts TOC structure from EPUB file
func (c *Api) extractTocFromEpub(id string) (map[string]interface{}, error) {
	var filepath string
	var err error

	// Use cache manager if available, otherwise fall back to old method
	if c.cacheManager != nil {
		filepath, err = c.cacheManager.GetOrExtractEpub(id)
	} else {
		filepath, err = c.getFileOrCache(id)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get EPUB file: %w", err)
	}

	book, err := epub.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open EPUB: %w", err)
	}
	defer book.Close()

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

	result := map[string]interface{}{
		"points":   p,
		"metadata": book.Opf.Metadata,
		"manifest": book.Opf.Manifest,
		"baseDir":  path.Dir(book.Container.Rootfile.Path),
	}

	return result, nil
}

// getBookContent 获取书籍内容（通过路径参数）
func (c *Api) getBookContent(r *gin.Context) {
	id := strings.TrimSuffix(r.Param("id"), ".epub")
	path1 := r.Param("path")

	_, err := c.getBookByIDV2(id)
	if err != nil {
		r.JSON(http.StatusInternalServerError, err)
		return
	}

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

	err = unzipSource(filepath, destDir)
	if err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}
	r.FileFromFS(path1, http.Dir(destDir))
}

// getBookContentByQuery 通过 query 参数获取书籍内容
func (c *Api) getBookContentByQuery(r *gin.Context) {
	// 从 query 参数获取书籍 ID
	id := r.Query("id")
	if id == "" {
		r.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "缺少必需的参数: id",
		})
		return
	}

	// 获取文件路径参数（可选）
	filePath := r.Query("path")
	if filePath == "" {
		filePath = "OEBPS/content.opf" // 默认返回 OPF 文件
	}

	_, err := c.getBookByIDV2(id)
	if err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取书籍信息失败: " + err.Error(),
		})
		return
	}

	// 获取或缓存书籍文件
	filepath, err := c.getFileOrCache(id)
	if err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取书籍文件失败: " + err.Error(),
		})
		return
	}

	// 解压目录
	destDir := path.Join(c.baseDir, id)
	if !Exists(destDir) {
		os.MkdirAll(destDir, fs.ModePerm)
	}

	// 检查目录是否为空，如果为空则解压
	if Exists(destDir) {
		s, _ := ioutil.ReadDir(destDir)
		if len(s) == 0 {
			// 目录为空，需要解压
			err := unzipSource(filepath, destDir)
			if err != nil {
				r.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "解压书籍文件失败: " + err.Error(),
				})
				return
			}
		}
	} else {
		// 目录不存在，创建并解压
		os.MkdirAll(destDir, fs.ModePerm)
		err := unzipSource(filepath, destDir)
		if err != nil {
			r.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "解压书籍文件失败: " + err.Error(),
			})
			return
		}
	}

	// 返回指定路径的文件
	r.FileFromFS(filePath, http.Dir(destDir))
}

// getFile 获取书籍文件
func (c *Api) getFile(id string) (int64, io.ReadCloser, error) {
	size, reader, err := c.contentApi.GetBook(id, "library")
	return size, reader, err
}

// getFileOrCache 获取或缓存书籍文件
func (c *Api) getFileOrCache(id string) (string, error) {
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
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	defer f.Close()
	_, err = f.Write(b)
	return filename, err
}

// expansionTree 展开书籍目录树
func (c *Api) expansionTree(ori []epub.NavPoint) []epub.NavPoint {
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

// stringToInt64 converts string to int64, returns 0 on error
func stringToInt64(s string) int64 {
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return val
}
