package StatsIO

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/sa-kemper/peertubestats/internal/LogHelp"
	"github.com/sa-kemper/peertubestats/pkg/peertubeApi"
)

type VideoStat struct {
	Time  time.Time `json:"time"`
	Likes Stat      `json:"likes"`
	Views Stat      `json:"views"`
}

type Stat struct {
	StartPercentage float64 `json:"start_percentage"`
	EndPercentage   float64 `json:"end_percentage"`
	Data            int64   `json:"data"`
}

// requestTimestamp will resolve a reasonable VideoStat for the given available ones and the requested one
// It throws an error on critical issues e.g. the whole year not being available or the years object is invalid
func requestTimestamp(ts time.Time, id int64) (result VideoStat, err error) {
	var lookupResult LikeView
	dll, found := Database.TimeSeriesDB.Video.Load(id)
	if !found {
		return fallbackRequestTimestamp(ts, id)
	}
	doubleLinkedListValue, ok := dll.(*DoubleLinkedList)
	if !ok {
		LogHelp.NewLog(LogHelp.Fatal, "cannot load double linked list from time series database", map[string]string{"id": strconv.FormatInt(id, 10), "timestamp": ts.Format(time.RFC3339)}).Log()
		// return will not be reached.
		return fallbackRequestTimestamp(ts, id)
	}

	lookupResult = lookupTimeSeriesSingle(doubleLinkedListValue, ts)
	return VideoStat{
		Time: ts,
		Likes: Stat{
			Data: lookupResult.Likes,
		},
		Views: Stat{
			Data: lookupResult.Views,
		},
	}, nil
}

func fallbackRequestTimestamp(ts time.Time, id int64) (result VideoStat, err error) {

	if ts.IsZero() {
		return VideoStat{}, errors.New("requestTimestamp called, but no timestamp provided")
	}
	// handle pre-recording date
	if ts.Before(Database.firstDataAvailable) {
		return VideoStat{
			Time:  ts,
			Likes: Stat{Data: 0},
			Views: Stat{Data: 0},
		}, nil
	}
	// handle pre video creation and post video deletion
	val, _ := Database.data.Load(id)
	metadata := val.(peertubeApi.VideoData)

	result, err = preCreationPostDeletionShortcut(ts, metadata)
	if !result.Time.IsZero() { // the timestamp was found and was returned
		return
	}

	// handle the base case, the stat is recorded and healthy.
	result, found := getStatOfDate(ts, id)
	if found {
		return result, nil
	}

	if _, err := os.Stat(filepath.Join(Database.DataFolder, ts.Format("2006"))); os.IsNotExist(err) {
		return VideoStat{}, errors.New("the requested year is not available")
	}

	return VideoStat{Time: ts}, nil
}

func preCreationPostDeletionShortcut(ts time.Time, metadata peertubeApi.VideoData) (stat VideoStat, err error) {
	// requestTimestamp was unaware of the publishing date. ts<publishDate
	if publishDate, err := metadata.GetPublishedAt(); err != nil || publishDate.IsZero() {
		return VideoStat{}, err
	} else if ts.Before(publishDate) {
		return VideoStat{
			Time: ts,
			Likes: Stat{
				StartPercentage: 0,
				EndPercentage:   0,
				Data:            0,
			},
			Views: Stat{
				StartPercentage: 0,
				EndPercentage:   0,
				Data:            0,
			},
		}, err
	}

	deletedVal, inDeletedDB := Database.deletedDb.Load(metadata.ID)
	if !inDeletedDB {
		return VideoStat{}, nil
	}
	if deletedVal.(time.Time).Before(ts) && !deletedVal.(time.Time).IsZero() {
		return VideoStat{
			Time: ts,
			Likes: Stat{
				StartPercentage: 0,
				EndPercentage:   0,
				Data:            metadata.Likes,
			},
			Views: Stat{
				StartPercentage: 0,
				EndPercentage:   0,
				Data:            metadata.Views,
			},
		}, nil
	}
	return VideoStat{}, nil
}

func getStatOfDate(ts time.Time, id int64) (result VideoStat, found bool) {
	if _, err := os.Stat(getRawFilePath(ts)); !os.IsNotExist(err) {
		videos := readRawResponses(ts)
		if len(videos) < 1 {
			// cannot read data
			LogHelp.NewLog(LogHelp.Error, "stat data was either not processed or is malformed", map[string]interface{}{"requestTimestamp": ts, "id": id}).Log()
			return
		}

		for _, video := range videos {
			if video.ID != id {
				continue
			}
			result = VideoStat{
				Time: ts,
				Likes: Stat{
					Data: video.Likes,
				},
				Views: Stat{
					Data: video.Views,
				},
			}
			found = true
			return // base case, everything works.
		}
	}
	return result, false
}
