package peertubeApi

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

func (api *ApiClient) GetVideoMetadata(id string) (data VideoData, err error) {
	const VideoMetadataEndpoint = "videos/{{id}}"
	//buffer := new(bytes.Buffer)
	endpointUrl := url.URL{
		Scheme: api.Protocol,
		Host:   api.Host,
		Path:   apiPrefix + strings.Replace(VideoMetadataEndpoint, "{{id}}", id, 1),
	}
	resp, err := api.doRequest(
		&http.Request{
			Method: http.MethodGet,
			URL:    &endpointUrl,
			Host:   api.Host,
		})
	if err != nil {
		println("error connecting to", endpointUrl.String())
		println(api.Host, api.Protocol, api.clientId)
		println(err.Error())
		return data, err
	}
	defer resp.Body.Close()
	//_, err = io.Copy(buffer, resp.Body)
	//if err != nil {
	//	println(err.Error())
	//	println(api.Host, api.Protocol, api.clientId)
	//}
	err = json.NewDecoder(resp.Body).Decode(&data)
	return data, err

}
