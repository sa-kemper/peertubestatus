package StatsIO

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path"
	"slices"
	"time"

	"github.com/sa-kemper/peertubestats/internal/LogHelp"
	"github.com/sa-kemper/peertubestats/pkg/peertubeapi"
)

func (statIO *StatsIO) ImportFromRaw(rawResponses [][]byte, serverVersion string, CollectionTime time.Time) (err error) {
	allResponses := []byte("# Peertube API Version: " + serverVersion + "\r\n")
	for _, response := range rawResponses {
		allResponses = append(allResponses, response...)
	}
	err = os.WriteFile(statIO.getRawFilePath(CollectionTime), allResponses, 0600)
	if err != nil {
		return errors.Join(errors.New("failed to write raw stats"), err)
	}
	go statIO.processRawImport(CollectionTime)
	go statIO.aggregateData(CollectionTime)
	return err
}

func (statIO *StatsIO) getRawFilePath(collectionTime time.Time) string {
	return path.Join(statIO.DataFolder, collectionTime.Format("2006"), collectionTime.Format("01"), collectionTime.Format("02")+".json")
}

func (statIO *StatsIO) processRawImport(collectionTime time.Time) {
	videos := statIO.ReadRawResponses(collectionTime)
	var videosDb = make(VideoDatabase)
	for id, video := range videos {
		videosDb[int64(id)] = video
	}

	err := statIO.mergeVideoDB(videosDb, &collectionTime)
	LogHelp.LogOnError("failed to merge input database into stored database", map[string]string{"collectionTime": collectionTime.Format("2006.01.02")}, err)
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

func (statIO *StatsIO) ReadRawResponses(collectionTime time.Time) (Videos []peertubeapi.VideoData) {
	Videos = make([]peertubeapi.VideoData, 0)

	VideosBytes, err := os.ReadFile(statIO.getRawFilePath(collectionTime))
	LogHelp.LogOnError("cannot read imported data", map[string]interface{}{"collectionTime": collectionTime.Format("2006.01.02")}, err)
	versionIndex := bytes.IndexByte(VideosBytes, byte('\n'))
	if versionIndex == -1 {
		LogHelp.NewLog(LogHelp.Error, "cannot find version header of raw data", map[string]interface{}{"collectionTime": collectionTime.Format("2006.01.02")}).Log()
		versionIndex = 0
	}
	decoder := json.NewDecoder(bytes.NewReader(VideosBytes[versionIndex+1:]))
	for {
		var video peertubeapi.VideoResponse
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
