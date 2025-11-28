package calibre

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jianyun8023/calibre-api/pkg/log"
)

// getIsbn 通过 ISBN 获取元数据
func (c *Api) getIsbn(c2 *gin.Context) {
	isbn := c2.Param("isbn")
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
	var jsonData map[string]interface{}
	resp, err := c.http.R().SetResult(&jsonData).SetQueryParam("q", query).Get(c.config.Metadata.DoubanUrl + "/v2/book/search")
	log.Infof("%s", resp.Request.URL)
	if err != nil {
		c2.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c2.JSON(http.StatusOK, resp.Result())
}
