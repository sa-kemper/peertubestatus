package StatsIO

import (
	"encoding/json"
	"os"
	"time"
)

type deletedDB map[int64]DeletedVideo
type DeletedVideo struct {
	Id      int64     `json:"id"`
	Deleted time.Time `json:"deleted"`
}

func loadDeletedDB() (vidDB deletedDB, err error) {
	dbBytes, err := os.ReadFile(Database.DataFolder + "/deleted.json")
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(dbBytes, &vidDB)
	if err != nil {
		return nil, err
	}
	return
}

func saveDeletedDB(db deletedDB) error {
	dbBytes, err := json.Marshal(db)
	if err != nil {
		return err
	}
	return os.WriteFile(Database.DataFolder+"/deleted.json", dbBytes, 0600)
}

func (db *VideoDatabase) VideoDelete(videoId int64, deletedTimestamp *time.Time) (err error) {
	if deletedTimestamp == nil {
		now := time.Now()
		deletedTimestamp = &now
	}

	deletedEntries, err := loadDeletedDB()
	if err != nil {
		return err
	}
	deletedEntries[videoId] = DeletedVideo{
		Id:      videoId,
		Deleted: *deletedTimestamp,
	}
	err = saveDeletedDB(deletedEntries)

	return err
}
