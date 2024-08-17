package content

import (
	"errors"
	"fmt"
	"github.com/jianyun8023/calibre-api/pkg/client"
	"github.com/jianyun8023/calibre-api/pkg/log"
	"github.com/spf13/cast"
	"io"
	"net/url"
	"strconv"
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

func (a *Api) GetAllBooksIds() ([]int64, error) {
	///ajax/search/library?num=10&offset=0&sort=id&sort_order=desc&query
	var data map[string]interface{}
	resp, err := a.R().SetResult(&data).
		SetQueryParam("num", "9999999").
		SetQueryParam("offset", "0").
		SetQueryParam("sort", "id").
		SetQueryParam("sort_order", "asc").
		SetQueryParam("query", "").
		Get("/ajax/search/library")
	log.Infof(resp.Request.URL + " " + resp.Status())
	if err != nil {
		return nil, err
	}
	bookIds := make([]int64, 0)
	bookIdsInterface := data["book_ids"].([]interface{})
	for _, id := range bookIdsInterface {
		bookIds = append(bookIds, int64(id.(float64)))
	}
	return bookIds, nil
}

func (a *Api) GetAllPublisher() ([]string, error) {
	///interface-data/field-names/publisher?library_id=library
	var publishers []string
	resp, err := a.R().SetResult(&publishers).
		SetQueryParam("library_id", "library").
		Get("/interface-data/field-names/publisher")
	log.Infof(resp.Request.URL + " " + resp.Status())
	return publishers, err
}

func (a *Api) GetBookMetaDatas(ids []int64, library string) ([]Book, error) {
	///cdb/cmd/list/0
	if library == "" {
		library = "library"
	}
	body := []interface{}{
		[]string{
			"id",
			"title",
			"authors",
			"comments",
			"size",
			"publisher",
			"pubdate",
			"isbn",
			"tags",
			"rating",
			"identifiers",
			"languages",
		},
		"id",
		"True",
		"id:>=" + strconv.FormatInt(ids[0], 10) + " and id:<=" + strconv.FormatInt(ids[len(ids)-1], 10),
		-1,
	}

	var data map[string]interface{}
	resp, err := a.R().SetResult(&data).SetBody(body).Post("/cdb/cmd/list/0")
	log.Infof(resp.Request.URL + " " + resp.Status())
	if err != nil {
		return nil, err
	}

	if data["err"] != nil {
		errMsg := data["err"].(string)
		log.Warn("error: " + errMsg)
		return nil, errors.New("error: " + errMsg)
	}

	data = data["result"].(map[string]interface{})

	books := make([]Book, 0)
	bookIdsInterface := data["book_ids"].([]interface{})

	bookData := data["data"].(map[string]interface{})
	titleMap := cast.ToStringMapString(bookData["title"])
	authorsMap := cast.ToStringMapStringSlice(bookData["authors"])
	commentsMap := cast.ToStringMapString(bookData["comments"])
	sizeMap := cast.ToStringMapInt64(bookData["size"])
	publisherMap := cast.ToStringMapString(bookData["publisher"])

	pubdateMap := bookData["pubdate"].(map[string]interface{})

	isbnMap := cast.ToStringMapString(bookData["isbn"])

	tagsMap := cast.ToStringMapStringSlice(bookData["tags"])
	ratingMap := cast.ToStringMap(bookData["rating"])

	identifiersMap := bookData["identifiers"].(map[string]interface{})
	languagesMap := cast.ToStringMapStringSlice(bookData["languages"])
	for _, id := range bookIdsInterface {
		book := Book{}
		book.ID = int64(id.(float64))
		strId := strconv.FormatInt(book.ID, 10)
		book.Title = titleMap[strId]
		book.Authors = authorsMap[strId]
		book.Comments = commentsMap[strId]
		book.Size = sizeMap[strId]
		book.Publisher = publisherMap[strId]
		m := pubdateMap[strId].(map[string]interface{})
		if m["v"] != nil {
			book.PubDate = cast.ToTime(m["v"])
		}
		book.Isbn = isbnMap[strId]
		book.Tags = tagsMap[strId]
		book.Rating = cast.ToFloat64(ratingMap[strId])
		book.Identifiers = cast.ToStringMapString(identifiersMap[strId])
		book.Languages = languagesMap[strId]
		books = append(books, book)
	}
	return books, nil
}

func convertIntMap(input map[string]interface{}) (map[string]int, error) {
	result := make(map[string]int)
	for k, v := range input {

		if v == nil {
			result[k] = 0
		} else {
			floatVal, ok := v.(float64)
			if !ok {
				return nil, fmt.Errorf("value for key %s is not of type float64", k)
			}
			result[k] = int(floatVal)
		}
	}
	return result, nil
}

func convertStringMap(input map[string]interface{}) (map[string]string, error) {

	result := make(map[string]string)
	for k, v := range input {
		if v == nil {
			result[k] = ""
			continue
		} else {
			strVal, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("value for key %s is not of type string", k)
			}
			result[k] = strVal
		}
	}
	return result, nil
}

func convertInt64Map(input map[string]interface{}) (map[string]int64, error) {

	result := make(map[string]int64)
	for k, v := range input {
		if v == nil {
			result[k] = 0
			continue
		} else {
			float64Val, ok := v.(float64)
			if !ok {
				return nil, fmt.Errorf("value for key %s is not of type int64", k)
			}
			result[k] = int64(float64Val)
		}
	}
	return result, nil

}
