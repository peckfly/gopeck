package netx

import (
	"net/url"
	"strings"
)

func ParseQueryMap(queryParams map[string]string) string {
	values := url.Values{}
	for key, value := range queryParams {
		values.Add(key, value)
	}
	queryString := values.Encode()
	return queryString
}

func ParseQuery(query string) map[string]string {
	values := make(map[string]string)
	params := strings.Split(query, "&")
	for _, param := range params {
		parts := strings.SplitN(param, "=", 2)
		if len(parts) == 2 {
			values[parts[0]] = parts[1]
		}
	}
	return values
}

// AddUrlParam add url param
func AddUrlParam(s string, paramMap map[string]string) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}
	params := url.Values{}
	for key, value := range paramMap {
		params.Add(key, value)
	}
	u.RawQuery = params.Encode()
	return u.String(), nil
}
