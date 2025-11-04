package StatsIO

import (
	"errors"
	"sync"

	"github.com/sa-kemper/peertubestats/pkg/peertubeApi"
)

func GetAllVideos() (Videos []peertubeApi.VideoData, err error) {
	var VideoDB *sync.Map
	if Database.data != nil {
		VideoDB = Database.data
	} else {
		VideoDB, err = loadVideoDB()
		if err != nil {
			return nil, err
		}
	}
	VideoDB.Range(
		func(k, v interface{}) bool {
			Videos = append(Videos, v.(peertubeApi.VideoData))
			return true
		})
	return Videos, nil
}

func GetVideo(id int64) (video peertubeApi.VideoData, err error) {
	var VideoDB *sync.Map
	if Database.data != nil {
		VideoDB = Database.data
	} else {
		VideoDB, err = loadVideoDB()
		if err != nil {
			return video, err
		}
	}
	vid, exists := VideoDB.Load(id)
	if !exists {
		return video, errors.New("video not found")
	}
	return vid.(peertubeApi.VideoData), nil
}
