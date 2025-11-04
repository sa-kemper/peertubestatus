package StatsIO

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/sa-kemper/peertubestats/internal/LogHelp"
)

type DeletedVideo struct {
	Id      int64     `json:"id"`
	Deleted time.Time `json:"deleted"`
}

func LoadDeletedDBFromDisk() (vidDB *sync.Map, err error) {
	vidDB = &sync.Map{}
	var DeletedDatabase = make(map[int64]DeletedVideo)

	handler, err := os.OpenFile(Database.DataFolder+"/deleted.json", os.O_RDONLY, 0600)
	LogHelp.LogOnError("cannot open deleted database file", nil, err)

	err = json.NewDecoder(handler).Decode(&DeletedDatabase)
	if err != nil {
		LogHelp.LogOnError("cannot decode deleted db from disk", nil, err)
		return
	}

	for _, DeletedItem := range DeletedDatabase {
		vidDB.Store(DeletedItem.Id, DeletedItem.Deleted)
	}
	return
}

func SaveDeletedDBToDisk(db *sync.Map) error {
	var DeletedDatabase map[int64]DeletedVideo
	db.Range(func(k, v interface{}) bool {
		DeletedDatabase[k.(int64)] = DeletedVideo{
			Id:      k.(int64),
			Deleted: v.(time.Time),
		}
		return true
	})
	handle, err := os.OpenFile(Database.DataFolder+"/deleted.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	err = json.NewEncoder(handle).Encode(DeletedDatabase)
	LogHelp.LogOnError("cannot encode DeletedDatabase to JSON", nil, err)

	return err
}

func VideoDelete(videoId int64, deletedTimestamp time.Time, deletedDB *sync.Map) {
	deletedDB.Store(videoId, DeletedVideo{
		Id:      videoId,
		Deleted: deletedTimestamp,
	})
}
