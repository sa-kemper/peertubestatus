package StatsIO

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/sa-kemper/peertubestats/internal/LogHelp"
	"github.com/sa-kemper/peertubestats/pkg/peertubeApi"
)

func (statIO *StatsIO) ImportFromRaw(rawResponses [][]byte, serverVersion string, CollectionTime time.Time) (err error) {
	allResponses := []byte("# Peertube API Version: " + serverVersion + "\r\n")
	for _, response := range rawResponses {
		allResponses = append(allResponses, response...)
	}

	dataPath := getRawFilePath(CollectionTime)
	err = os.MkdirAll(path.Dir(dataPath), 0700)
	LogHelp.LogOnError("cannot create directory", map[string]string{"path": dataPath}, err)

	err = os.WriteFile(getRawFilePath(CollectionTime), allResponses, 0600)
	if err != nil {
		return errors.Join(errors.New("failed to write raw stats"), err)
	}
	statIO.processRawImport(CollectionTime)

	return err
}

/*
processRawImport validates the raw data, loads additional data, and updates the videoDB (if required)
It errors to the LogHelp utility, as it is meant to run concurrently.
*/
func (statIO *StatsIO) processRawImport(collectionTime time.Time) {
	videos := readRawResponses(collectionTime) // TODO: Adapt to stateless port
	var videosDb = sync.Map{}
	var LocalWg sync.WaitGroup
	LocalWg.Add(len(videos))

	for _, video := range videos {
		videosDb.Store(video.ID, video)
		go func() {
			defer LocalWg.Done()
			thumbnailPath := path.Join(Database.DataFolder, video.ThumbnailPath)
			mkErr := os.MkdirAll(path.Dir(thumbnailPath), 0700)
			if mkErr != nil {
				LogHelp.NewLog(LogHelp.Fatal, "cannot create directory", map[string]string{"path": path.Dir(thumbnailPath), "fullPath": thumbnailPath}).Log()
			}
			absPath, _ := filepath.Abs(thumbnailPath)
			// BUG(Samuel): if the collectionTime is far in the past, it is impossible to retrieve the original thumbnail. the current thumbnail is obtained regardless (if it has the same path). This may be subject to a fix in the future.
			if stat, _ := os.Stat(absPath); stat == nil { // if the file does not exist. load it.
				thumb, err := statIO.Api.GetThumbnail(video.ID)
				if err != nil {
					LogHelp.NewLog(LogHelp.Error, "cannot get thumbnail", map[string]string{"error": err.Error(), "localPath": thumbnailPath, "videoID": strconv.FormatInt(video.ID, 10), "videoThumbnailPath": video.ThumbnailPath}).Log()
					return // this stops execution for the thumbnail download.
				}
				fHandler, err := os.OpenFile(absPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
				LogHelp.LogOnError("cannot open thumbnail file", map[string]interface{}{"error": err, "localPath": thumbnailPath}, err)
				_, err = io.Copy(fHandler, bytes.NewBuffer(thumb))
				if err != nil {
					LogHelp.NewLog(LogHelp.Fatal, "cannot write thumbnail file", map[string]string{"error": err.Error()}).Log()
					return
				}
			}
		}()
	}

	var videoDBDebugObj = make(map[int64]peertubeApi.VideoData)
	videosDb.Range(func(k, v interface{}) bool {
		videoDBDebugObj[k.(int64)] = v.(peertubeApi.VideoData)
		return true
	})

	currentDB, err := loadVideoDB()
	LogHelp.LogOnError("cannot load video db", nil, err)
	deletedDB, err := LoadDeletedDBFromDisk()
	LogHelp.LogOnError("cannot load deleted db", nil, err)

	LocalWg.Wait()

	err = mergeVideoDB(currentDB, &videosDb, deletedDB, collectionTime)
	LogHelp.LogOnError("failed to merge input database into stored database", map[string]string{"collectionTime": collectionTime.Format("2006.01.02")}, err)

	err = SaveDeletedDBToDisk(deletedDB)
	LogHelp.LogOnError("failed to save deleted db to disk", nil, err)
	err = saveVideoDB(currentDB, time.Now())
	LogHelp.LogOnError("failed to save video db to disk", nil, err)

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

func readRawResponses(collectionTime time.Time) (Videos []peertubeApi.VideoData) {
	Videos = make([]peertubeApi.VideoData, 0)

	VideosBytes, err := os.ReadFile(getRawFilePath(collectionTime))
	if err != nil {
		LogHelp.LogOnError("cannot read imported data", map[string]interface{}{"collectionTime": collectionTime.Format("2006.01.02")}, err)
		return
	}
	versionIndex := bytes.IndexByte(VideosBytes, byte('\n'))
	if versionIndex == -1 {
		LogHelp.NewLog(LogHelp.Error, "cannot find version header of raw data", map[string]interface{}{"collectionTime": collectionTime.Format("2006.01.02")}).Log()
		versionIndex = 0
	}
	decoder := json.NewDecoder(bytes.NewReader(VideosBytes[versionIndex+1:]))
	for {
		var video peertubeApi.VideoResponse
		err = decoder.Decode(&video)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			LogHelp.LogOnError("error parsing imported data", map[string]interface{}{"collectionTime": collectionTime.Format("2006.01.02")}, err)
			return
		}
		Videos = slices.Concat(Videos, video.Data)
	}
	return
}

func getRawFilePath(collectionTime time.Time) (result string) {
	inputPath := path.Join(Database.DataFolder, collectionTime.Format("2006"), collectionTime.Format("01"), collectionTime.Format("02")+".json")
	abs, err := filepath.Abs(inputPath)
	if err == nil {
		return abs
	}
	return inputPath
}
