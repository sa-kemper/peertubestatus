package StatsIO

import (
	"encoding/json"
	"errors"
	"os"
	"path"

	"github.com/sa-kemper/peertubestats/pkg/peertubeApi"
)

// VideoMetadata queries the "DB" for the metadata of the video and returns it.
func VideoMetadata(id int64) (result peertubeApi.VideoData, err error) {
	Database.dataMutex.RLock()
	defer Database.dataMutex.RUnlock()
	if Database.data == nil {
		db, err := loadVideoDB()
		if err != nil {
			return result, err
		}
		Database.data = &db
	}
	result, ok := (*Database.data)[id]
	if !ok {
		return result, errors.New("requested video does not exist")
	}
	return result, nil
}

// loadVideoDB loads all metadata of every video ever seen.
func loadVideoDB() (VideoDatabase, error) {
	jsonDBBytes, err := os.ReadFile(path.Join(Database.DataFolder, "videoDB.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[int64]peertubeApi.VideoData), nil
		}
		return nil, err
	}
	var dbMap map[int64]peertubeApi.VideoData
	err = json.Unmarshal(jsonDBBytes, &dbMap)
	return dbMap, err
}
