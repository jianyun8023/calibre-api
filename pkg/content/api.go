package content

import (
	"github.com/jianyun8023/calibre-api/pkg/client"
	"github.com/jianyun8023/calibre-api/pkg/log"
	"io"
	"net/url"
	"strings"
)

type Api struct {
	*client.Client
}

func NewClient(baseUrl string) (Api, error) {

	parsedURL, err := url.Parse(baseUrl)
	if err != nil {
		return Api{}, err
	}
	api, err := client.New(&client.Config{
		Host:  parsedURL.Host,
		HTTPS: parsedURL.Scheme == "https",
	})
	api.BaseURL = baseUrl
	if err != nil {
		return Api{}, err
	}
	return Api{
		Client: api,
	}, nil
}

func (a *Api) DeleteBooks(bookIds []string, library string) error {
	if library == "" {
		library = "library"
	}
	///cdb/delete-books/264728/library
	ids := strings.Join(bookIds, ",")
	resp, err := a.R().SetPathParam("ids", ids).SetPathParam("library", library).Post("/cdb/delete-books/{ids}/{library}")
	log.Infof(resp.Request.URL + " " + resp.Status())
	return err
}

func (a *Api) UpdateMetaData(id string, metadata map[string]interface{}, library string) (map[string]Content, error) {
	if library == "" {
		library = "library"
	}
	///cdb/delete-books/264728/library
	body := map[string]interface{}{
		"changes": metadata,
	}

	var data map[string]Content
	resp, err := a.R().SetResult(&data).SetPathParam("id", id).SetPathParam("library", library).SetBody(body).Post("/cdb/set-fields/{id}/{library}")
	log.Infof(resp.Request.URL + " " + resp.Status())
	return data, err
}

func (a *Api) GetCover(id string, library string) (int64, io.ReadCloser, error) {
	if library == "" {
		library = "library"
	}
	///get/cover/269220/library
	resp, err := a.R().SetDoNotParseResponse(true).SetPathParam("id", id).SetPathParam("library", library).Get("/get/cover/{id}/{library}")
	if err != nil {

		return 0, nil, err
	}
	response := resp.RawResponse
	log.Infof(resp.Request.URL + " " + resp.Status())
	return response.ContentLength, response.Body, err
}

func (a *Api) GetBook(id string, library string) (int64, io.ReadCloser, error) {
	if library == "" {
		library = "library"
	}
	///get/EPUB/269220/library
	resp, err := a.R().SetDoNotParseResponse(true).SetPathParam("id", id).SetPathParam("library", library).Get("/get/EPUB/{id}/{library}")
	if err != nil {

		return 0, nil, err
	}
	response := resp.RawResponse
	log.Infof(resp.Request.URL + " " + resp.Status())
	return response.ContentLength, response.Body, err

}
