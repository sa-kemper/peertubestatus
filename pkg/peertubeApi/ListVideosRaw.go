package peertubeApi

import (
	"errors"
	"io"
	"net/http"
	"net/url"
)

var ErrorNoMoreResults = errors.New("no more results")

func (api *ApiClient) ListVideosRaw(args ListVideosParams) (data []byte, err error) {
	const endpoint = "videos"
	_, err = ValidateVideoIncludeFlags(args.Include, api.isAdmin)
	if err != nil {
		return
	}

	var listVideosUrl = url.URL{
		Scheme:     api.Protocol,
		Host:       api.Host,
		Path:       apiPrefix + endpoint,
		ForceQuery: true,
		RawQuery:   toQueryParams(args).Encode(),
	}
	listVideosUrl.Query().Add("host", api.Host)

	httpResponse, err := api.doRequest(&http.Request{
		Method: http.MethodGet,
		URL:    &listVideosUrl,
		Header: api.headers,
		Host:   api.Host,
	})
	if err != nil {
		return
	}

	defer httpResponse.Body.Close()
	data, err = io.ReadAll(httpResponse.Body)
	if len(data) < 25 {
		return []byte{}, ErrorNoMoreResults
	}
	return

}
