package lanzou

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jianyun8023/calibre-api/pkg/client"
	"github.com/jianyun8023/calibre-api/pkg/lanzou"
	"net/http"
	"net/url"
)

type Api struct {
	client *lanzou.Lanzou
}

func (c Api) SetupRouter(r *gin.Engine) {
	r.GET("/api/lanzou", c.parse)
}

func NewClient() (Api, error) {
	lanzouAPI, err := lanzou.New(&client.Config{})
	if err != nil {
		return Api{}, err
	}
	return Api{
		client: lanzouAPI,
	}, nil
}

func (c Api) parse(r *gin.Context) {
	lanzouUrl := r.Query("url")
	pwd := r.Query("pwd")

	_, err := url.Parse(lanzouUrl)
	if err != nil {
		r.JSON(http.StatusBadRequest,
			&Response{
				Msg:  fmt.Sprintf("%v", err),
				Code: 401,
			})
		return
	}

	link, err := c.client.ResolveShareURL(lanzouUrl, pwd)
	if err != nil {
		r.JSON(http.StatusInternalServerError,
			&Response{
				Msg:  fmt.Sprintf("%v", err),
				Code: 500,
			})
		return
	}
	r.JSON(http.StatusOK, &Response{
		Msg:  "ok",
		Code: 200,
		Data: link,
	})
}

type Response struct {
	Code int64                 `json:"code"`
	Data []lanzou.ResponseData `json:"data"`
	Msg  string                `json:"msg"`
}
