package StatsIO

import (
	"github.com/sa-kemper/peertubestats/pkg/peertubeapi"
)

func GetAllVideos() (Videos []peertubeapi.VideoData, err error) {
	var VideoDB VideoDatabase
	VideoDB, err = loadVideoDB()
	if err != nil {
		return nil, err
	}
	for _, video := range VideoDB {
		Videos = append(Videos, video)
	}
	return Videos, nil
}

func GetVideo(id int64) (video peertubeapi.VideoData, err error) {
	var VideoDB VideoDatabase
	VideoDB, err = loadVideoDB()
	if err != nil {
		return
	}
	return VideoDB[id], nil
}
