package peertubeApi

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// GetThumbnail is a utility function that calls GetVideoMetadata and obtains the provided thumbnail from that response
// GetThumbnail takes the video id on peertube
func (api *ApiClient) GetThumbnail(id int64) (thumbnailData []byte, err error) {
	var buffer bytes.Buffer
	videoMetadata, err := api.GetVideoMetadata(strconv.FormatInt(id, 10))
	if err != nil {
		return
	}
	if videoMetadata.ThumbnailPath == "" {
		return
	}
	endpointUrl := url.URL{
		Scheme: api.Protocol,
		Host:   api.Host,
		Path:   videoMetadata.ThumbnailPath,
	}
	response, err := api.doRequest(
		&http.Request{
			Method: http.MethodGet,
			URL:    &endpointUrl,
			Proto:  api.Protocol,
			Host:   api.Host,
		})
	if err != nil {
		println("error connecting to", endpointUrl.String())
		println(api.Host, api.Protocol, api.clientId)
		println(err.Error())

		return
	}
	defer response.Body.Close()

	_, err = io.Copy(&buffer, response.Body)
	return buffer.Bytes(), err
}
