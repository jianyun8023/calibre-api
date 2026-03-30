package douban

import (
	"net/url"
	"regexp"
)

var (
	IdPattern     = regexp.MustCompile(`.*/subject/(\d+)/?`)
	SeriesPattern = regexp.MustCompile(`.*/series/(\d+)/?`)
	TagsPattern   = regexp.MustCompile(`criteria = '(.+)'`)
)

func ParseQuery(query string) map[string]string {
	values, err := url.ParseQuery(query)
	if err != nil {
		return make(map[string]string)
	}
	result := make(map[string]string)
	for k, v := range values {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}

func IsBookUrl(u string) bool {
	return IdPattern.MatchString(u)
}
