package main

import "time"

type Book struct {
	AuthorSort  string   `json:"author_sort"`
	Authors     string   `json:"authors"`
	Cover       string   `json:"cover"`
	Formats     []string `json:"formats"`
	ID          int64    `json:"id"`
	Identifiers struct {
		MobiAsin string `json:"mobi-asin"`
	} `json:"identifiers"`
	Isbn         string    `json:"isbn"`
	Languages    []string  `json:"languages"`
	LastModified time.Time `json:"last_modified"`
	Pubdate      time.Time `json:"pubdate"`
	Publisher    string    `json:"publisher"`
	SeriesIndex  float64   `json:"series_index"`
	Size         int64     `json:"size"`
	Tags         []string  `json:"tags"`
	Timestamp    time.Time `json:"timestamp"`
	Title        string    `json:"title"`
	UUID         string    `json:"uuid"`
}
