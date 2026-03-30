package calibre

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jianyun8023/calibre-api/pkg/log"
)

// getIsbn 通过 ISBN 获取元数据
func (c *Api) getIsbn(c2 *gin.Context) {
	isbn := c2.Param("isbn")

	// 本地模式：直接调用 douban 服务
	if c.doubanMode == "local" && c.doubanService != nil {
		result := c.doubanService.Get(isbn, c.config.Metadata.DoubanConfig.IsbnUrl)
		if !result.Success || len(result.Books) == 0 {
			log.Warnf("ISBN %s not found in douban local service", isbn)
			c2.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
			return
		}
		log.Infof("ISBN %s found in douban local service", isbn)
		c2.JSON(http.StatusOK, result.Books[0])
		return
	}

	// HTTP 降级模式
	var jsonData map[string]interface{}
	resp, err := c.http.R().SetResult(&jsonData).Get(c.config.Metadata.DoubanUrl + "/v2/book/isbn/" + isbn)
	log.Infof("%s", resp.Request.URL)
	if err != nil {
		c2.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c2.JSON(http.StatusOK, resp.Result())
}

// queryMetadata 查询元数据
func (c *Api) queryMetadata(c2 *gin.Context) {
	query := c2.Query("query")

	// 本地模式：直接调用 douban 服务
	if c.doubanMode == "local" && c.doubanService != nil {
		result := c.doubanService.Search(query, 5) // 默认返回 5 个结果
		if !result.Success {
			log.Warnf("Query '%s' failed in douban local service", query)
			c2.JSON(http.StatusNotFound, gin.H{"error": "no results found"})
			return
		}
		log.Infof("Query '%s' found %d results in douban local service", query, result.Count)
		c2.JSON(http.StatusOK, result)
		return
	}

	// HTTP 降级模式
	var jsonData map[string]interface{}
	resp, err := c.http.R().SetResult(&jsonData).SetQueryParam("q", query).Get(c.config.Metadata.DoubanUrl + "/v2/book/search")
	log.Infof("%s", resp.Request.URL)
	if err != nil {
		c2.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c2.JSON(http.StatusOK, resp.Result())
}
