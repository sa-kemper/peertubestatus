package StatsIO

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"slices"
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
	var videosDb = make(VideoDatabase)
	var LocalWg sync.WaitGroup
	LocalWg.Add(len(videos))
	for id, video := range videos {
		videosDb[int64(id)] = video
		go func() {
			defer LocalWg.Done()
			thumbnailPath := path.Join(Database.DataFolder, video.ThumbnailPath)
			mkErr := os.MkdirAll(path.Dir(thumbnailPath), 0700)
			if mkErr != nil {
				LogHelp.NewLog(LogHelp.Fatal, "cannot create directory", map[string]string{"path": path.Dir(thumbnailPath), "fullPath": thumbnailPath})
			}
			absPath, _ := filepath.Abs(thumbnailPath)

			if stat, _ := os.Stat(thumbnailPath); stat == nil {
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

	err := statIO.mergeVideoDB(videosDb, &collectionTime)
	LogHelp.LogOnError("failed to merge input database into stored database", map[string]string{"collectionTime": collectionTime.Format("2006.01.02")}, err)
	LocalWg.Wait()
}

func (statIO *StatsIO) mergeVideoDB(inputDatabase VideoDatabase, recordedTs *time.Time) (err error) {
	statIO.dataMutex.Lock()
	defer statIO.dataMutex.Unlock()

	if recordedTs == nil {
		now := time.Now()
		recordedTs = &now
	}

	for i := range *statIO.data {
		if _, ok := inputDatabase[i]; !ok {
			DbVideo := (*(*statIO).data)[i]
			deleted := statIO.data.VideoIsDeleted(DbVideo)
			if deleted == nil {
				LogHelp.NewLog(LogHelp.Error, "cannot retrieve deleted state of video", map[string]interface{}{"video": inputDatabase[i]})
				continue
			}

			if *deleted {
				err = statIO.data.VideoDelete(i, recordedTs)
				if err != nil {
					return err
				}
			}
		}
	}
	for i := range inputDatabase {
		if !statIO.data.VideoExists(inputDatabase[i]) {
			statIO.data.VideoAdd(inputDatabase[i])
		}
	}

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

func (statIO *StatsIO) aggregateData(time time.Time) {
	// TODO: Implelemt
	panic("implement me")
}
