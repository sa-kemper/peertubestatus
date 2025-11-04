package StatsIO

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/sa-kemper/peertubestats/internal/LogHelp"
	"github.com/sa-kemper/peertubestats/pkg/peertubeApi"
)

// loadVideoDB loads all metadata of every video ever seen.
func loadVideoDB() (result *sync.Map, err error) {
	result = new(sync.Map)
	jsonDBBytes, err := os.ReadFile(path.Join(Database.DataFolder, "videoDB.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return

	}
	var dbMap map[int64]peertubeApi.VideoData
	err = json.Unmarshal(jsonDBBytes, &dbMap)
	if err != nil {
		return
	}

	for _, videoData := range dbMap {
		result.Store(videoData.ID, videoData)
	}
	return result, err
}

// mergeVideoDB adds the inputDatabase to the currentData, while adding removed entries to the deletedDb.
// mergeVideoDB does not remove entries from the currentData as it is still used for metadata lookup.
// CATION: Call Load AND Save the Deleted db before operating on it using the LoadDeletedDBFromDisk and SaveDeletedDBToDisk functions respectively
func mergeVideoDB(currentData *sync.Map, inputDatabase *sync.Map, deletedDb *sync.Map, recordedTs time.Time) (err error) {
	// Check for deleted or modified videos
	currentData.Range(func(key, value interface{}) bool {
		// TODO: Swap the entries if the thumbnail path changes.
		// Why? because they are regenerated from time to time, and this results in broken thumbnail requests.

		// if the key from the input is not found, it is either deleted or new
		if _, found := inputDatabase.Load(key); !found {
			videoFromDB := value.(peertubeApi.VideoData)

			// Retrieve deletion time for the video
			deletedValue, foundInDeletedDB := deletedDb.Load(videoFromDB.ID)

			if foundInDeletedDB {
				deletionTs := deletedValue.(time.Time)

				// If video was foundInDeletedDB before the recorded state, ignore
				if deletionTs.Before(recordedTs) {
					return true
				}

				// If video is foundInDeletedDB in the future, log a warning
				if deletionTs.After(recordedTs) {
					LogHelp.NewLog(LogHelp.Warn, "Writing video data in the past.", map[string]string{
						"VideoID":      strconv.FormatInt(videoFromDB.ID, 10),
						"DataRecorded": recordedTs.Format(time.RFC3339),
						"VideoDeleted": deletionTs.Format(time.RFC3339),
					}).Log()

					// Update deletion timestamp
					deletedDb.Store(videoFromDB.ID, recordedTs)
				}
				return true
			}
			if !foundInDeletedDB {
				// video is not found in the new data, and it is not in the foundInDeletedDB
				VideoDelete(videoFromDB.ID, recordedTs, deletedDb)
			}
		}

		return true
	})

	// Merge input database into current data
	inputDatabase.Range(func(key, value interface{}) bool {
		keyint, ok1 := key.(int64)
		LogHelp.ErrorOnNotOK("cannot cast inputdatabase index to int64 (mergeVideoDB)", nil, ok1)
		valueVideo, ok2 := value.(peertubeApi.VideoData)
		LogHelp.ErrorOnNotOK("cannot cast inputdatabase value to peertubeApi.VideoData (mergeVideoDB)", nil, ok2)
		currentData.Store(keyint, valueVideo)
		return ok1 && ok2
	})

	return nil
}

func saveVideoDB(Db *sync.Map, ts time.Time) error {
	var fileDB = make(map[int64]peertubeApi.VideoData)
	Db.Range(func(k, v interface{}) (ok bool) {
		fileDB[k.(int64)], ok = v.(peertubeApi.VideoData)
		LogHelp.ErrorOnNotOK("cannot add key value pair to map", nil, ok)
		return ok
	})
	monthHandle, err := os.OpenFile(Database.DataFolder+string(os.PathSeparator)+ts.Format("2006")+string(os.PathSeparator)+ts.Format("1")+".json", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer monthHandle.Close()

	yearHandle, err := os.OpenFile(Database.DataFolder+string(os.PathSeparator)+ts.Format("2006")+".json", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer yearHandle.Close()

	fullHandle, err := os.OpenFile(Database.DataFolder+string(os.PathSeparator)+"videoDB.json", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer fullHandle.Close()

	byts, err := json.Marshal(fileDB)
	if err != nil {
		return err
	}

	var monthErr, yearErr, fullErr error
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()

		_, monthErr = monthHandle.Write(bytes.Clone(byts))
	}()
	go func() {
		defer wg.Done()
		_, yearErr = yearHandle.Write(bytes.Clone(byts))
	}()
	go func() {
		defer wg.Done()
		_, fullErr = fullHandle.Write(bytes.Clone(byts))
	}()
	wg.Wait()
	if monthErr != nil || yearErr != nil || fullErr != nil {
		return errors.Join(errors.New("failed to save video db"), monthErr, yearErr, fullErr)
	}
	return nil
}

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
