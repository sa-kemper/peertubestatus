package StatsIO

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
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

	dataPath := statIO.getRawFilePath(CollectionTime)
	err = os.MkdirAll(path.Dir(dataPath), 0700)
	LogHelp.LogOnError("cannot create directory", map[string]string{"path": dataPath}, err)

	err = os.WriteFile(statIO.getRawFilePath(CollectionTime), allResponses, 0600)
	if err != nil {
		return errors.Join(errors.New("failed to write raw stats"), err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go statIO.processRawImport(CollectionTime, &wg)
	//go statIO.aggregateData(CollectionTime)
	wg.Wait()
	return err
}

func (statIO *StatsIO) getRawFilePath(collectionTime time.Time) (result string) {
	inputPath := path.Join(statIO.DataFolder, collectionTime.Format("2006"), collectionTime.Format("01"), collectionTime.Format("02")+".json")
	abs, err := filepath.Abs(inputPath)
	if err == nil {
		return abs
	}
	return inputPath
}

func (statIO *StatsIO) processRawImport(collectionTime time.Time, wg *sync.WaitGroup) {
	defer wg.Done()
	videos := statIO.ReadRawResponses(collectionTime)
	var videosDb = sync.Map{}
	var LocalWg sync.WaitGroup
	LocalWg.Add(len(videos))
	for id, video := range videos {
		videosDb.Store(id, video)
		go func() {
			defer LocalWg.Done()
			thumbnailPath := path.Join(Database.DataFolder, video.ThumbnailPath)
			bits, _ := json.Marshal(thumbnailPath)
			if strings.Contains(string(bits), "6aT1w9gQWwD3bTvqZZBPu2") {
				print("Debug entry point\n")
				print(video.ThumbnailPath)
			}

			mkErr := os.MkdirAll(path.Dir(thumbnailPath), 0700)
			if mkErr != nil {
				LogHelp.NewLog(LogHelp.Fatal, "cannot create directory", map[string]string{"path": path.Dir(thumbnailPath), "fullPath": thumbnailPath}).Log()
			}
			absPath, _ := filepath.Abs(thumbnailPath)

			if stat, _ := os.Stat(thumbnailPath); stat == nil {
				// we do not use the API client as we know most of the video's metadata and creating a new api client just for this would be an overhead.
				client := http.Client{
					Timeout: time.Second * 5,
				}
				response, err := client.Do(
					&http.Request{
						Method: http.MethodGet,
						URL: &url.URL{
							Scheme: "https",
							Host:   flag.Lookup("api-host").Value.String(),
							Path:   video.ThumbnailPath,
						},
						Host: flag.Lookup("api-host").Value.String(),
					},
				)
				if err != nil || response.StatusCode != http.StatusOK {
					var body []byte
					if response != nil {
						body, _ = io.ReadAll(response.Body)
					}
					LogHelp.NewLog(LogHelp.Fatal, "cannot request thumbnail", map[string]string{"error": err.Error(), "response": string(body)}).Log()
					return
				}
				defer response.Body.Close()
				fHandler, err := os.OpenFile(absPath, os.O_CREATE|os.O_RDWR, 0600)
				if err != nil || response.StatusCode != http.StatusOK {
					LogHelp.NewLog(LogHelp.Fatal, "cannot create thumbnail file", map[string]string{"error": err.Error()}).Log()
					return
				}
				_, err = io.Copy(fHandler, response.Body)
				if err != nil {
					LogHelp.NewLog(LogHelp.Fatal, "cannot write thumbnail file", map[string]string{"error": err.Error()}).Log()
					return
				}
			}
		}()
	}

	err := statIO.mergeVideoDB(&videosDb, &collectionTime)
	LogHelp.LogOnError("failed to merge input database into stored database", map[string]string{"collectionTime": collectionTime.Format("2006.01.02")}, err)
	LocalWg.Wait()
}

// mergeVideoDB adds the inputDatabase to the currentData, while adding removed entries to the deletedDb.
// mergeVideoDB does not remove entries from the currentData as it is still used for metadata lookup.
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

func (statIO *StatsIO) ReadRawResponses(collectionTime time.Time) (Videos []peertubeApi.VideoData) {
	Videos = make([]peertubeApi.VideoData, 0)

	VideosBytes, err := os.ReadFile(statIO.getRawFilePath(collectionTime))
	LogHelp.LogOnError("cannot read imported data", map[string]interface{}{"collectionTime": collectionTime.Format("2006.01.02")}, err)
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

// paddNumber is used to get the file name part of the saved data, we use time.Format("02") which is a zero padded date. this helps retrieving the file.
func paddNumber(width int, number int) string {
	numStr := strconv.Itoa(number)
	for len(numStr) < width {
		numStr = "0" + numStr
	}
	return numStr
}

func (statIO *StatsIO) aggregateData(time time.Time) (monthDb, YearDb, FullDb map[int64]peertubeApi.VideoData) {
	var db = new(sync.Map)
	monthDb, YearDb, FullDb = make(map[int64]peertubeApi.VideoData), make(map[int64]peertubeApi.VideoData), make(map[int64]peertubeApi.VideoData)

	err := aggregateMonth(time, db)
	LogHelp.LogOnError("error aggregating month data", map[string]interface{}{"collectionTime": time.Format("2006.01.02")}, err)

	db.Range(func(key, value interface{}) bool { // copy the state of the database into a separate object.
		monthDb[key.(int64)] = value.(peertubeApi.VideoData)
		return true
	})

	err = aggregateYear(time, db)
	LogHelp.LogOnError("error aggregating year data", map[string]interface{}{"collectionTime": time.Format("2006.01.02")}, err)

	db.Range(func(key, value interface{}) bool { // copy the state of the database into a separate object.
		YearDb[key.(int64)] = value.(peertubeApi.VideoData)
		return true
	})

	err = aggregateFull(db)
	LogHelp.LogOnError("error aggregating full data", map[string]interface{}{"collectionTime": time.Format("2006.01.02")}, err)

	db.Range(func(key, value interface{}) bool { // copy the state of the database into a separate object.
		FullDb[key.(int64)] = value.(peertubeApi.VideoData)
		return true
	})
	return
}

func aggregateFull(db *sync.Map) error {
	wg := sync.WaitGroup{}
	yearsToCollect := time.Now().Year() - Database.firstDataAvailable.Year()
	wg.Add(yearsToCollect)
	for iterator := 0; iterator < yearsToCollect; iterator++ {
		yearDatabaseBasePath := path.Join(Database.DataFolder, paddNumber(2, iterator)+".json")
		yearDatabaseAbsPath, err := filepath.Abs(yearDatabaseBasePath)
		if err != nil {
			return err
		}
		go func() {
			defer wg.Done()
			fHandler, localErr := os.OpenFile(yearDatabaseAbsPath, os.O_RDONLY, 0600)
			if localErr != nil {
				err = errors.Join(err, localErr)
				return
			}
			defer fHandler.Close()
			var responses []peertubeApi.VideoResponse
			decodeErr := json.NewDecoder(fHandler).Decode(&responses)
			if decodeErr != nil {
				err = errors.Join(err, decodeErr)
				return
			}
			aggregate(responses, db)
		}()
	}
	wg.Wait()
	return nil
}

func aggregateYear(timeStamp time.Time, db *sync.Map) (err error) {
	wg := sync.WaitGroup{}
	wg.Add(int(timeStamp.Month()))
	for iterator := 0; iterator < int(timeStamp.Month()); iterator++ {
		monthDatabasePath := path.Join(Database.DataFolder, timeStamp.Format("2006"), paddNumber(2, iterator)+".json")
		go func() {
			defer wg.Done()
			fHandler, LocalErr := os.OpenFile(monthDatabasePath, os.O_RDONLY, 0600)
			if LocalErr != nil {
				err = errors.Join(err, LocalErr)
				return
			}
			var responses []peertubeApi.VideoResponse
			err = json.NewDecoder(fHandler).Decode(&responses)
			if err != nil {
				err = errors.Join(err, LocalErr)
				return
			}
			aggregate(responses, db)
		}()
	}

	wg.Wait()
	return
}

func aggregateMonth(time time.Time, db *sync.Map) error {
	folder := path.Join(Database.DataFolder, time.Format("2006"), time.Format("01"))
	abs, _ := filepath.Abs(folder)
	err := filepath.WalkDir(abs, func(path string, d fs.DirEntry, err error) error {
		var responses []peertubeApi.VideoResponse
		fHandle, _ := os.OpenFile(path, os.O_RDONLY, 0600)
		err = json.NewDecoder(fHandle).Decode(&responses)
		if err != nil {
			return err
		}
		aggregate(responses, db)
		return nil
	})
	return err
}

// aggregate is simple, go through the data, if it seems more up to date, swap the db entry.
// the up-to-date ness is determined by the likes and views which are never decreasing.
func aggregate(responses []peertubeApi.VideoResponse, db *sync.Map) {
	var responseMap map[int64]peertubeApi.VideoData
	for _, response := range responses {
		for _, video := range response.Data {
			responseMap[video.ID] = video
			vid, notFound := db.Load(video.ID)
			dbVideo := vid.(peertubeApi.VideoData)
			if notFound {
				db.Store(video.ID, video)
				continue
			}
			if dbVideo.Views < video.Views || dbVideo.Likes < video.Likes {
				db.Swap(video.ID, video)
				// TODO: load the new thumbnail, shadow the old one?
				// Details: sometimes the thumbnail path changes, and maybe the new one gets saved by the other processing functions, but idk what to do with the old ones yet, this is yet to be implemented.
			}
		}
	}
	// walk through the responses, if a video has been deleted, delete it from the db as well.
	db.Range(func(key, value interface{}) bool {
		_, found := responseMap[key.(int64)]
		if !found {
			db.Delete(key)
		}
		return true
	})

}
