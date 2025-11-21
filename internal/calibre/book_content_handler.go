package calibre

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
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
func (c *Api) getBookToc(r *gin.Context) {
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

// getBookContent 获取书籍内容（通过路径参数）
func (c *Api) getBookContent(r *gin.Context) {
	id := strings.TrimSuffix(r.Param("id"), ".epub")
	path1 := r.Param("path")

	_, err := c.getBookByID(id)
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

	_, err := c.getBookByID(id)
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
