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

	// The data is not available. AUTO REPAIR!

	// handle pre video creation and post video deletion
	metadata, err := VideoMetadata(id)
	if err != nil {
		// video was not found
		return VideoStat{}, err
	}

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

	dltdb, err := loadDeletedDB()
	if err != nil {
		return VideoStat{}, err
	}
	if videoDeleted, wasDeleted := dltdb[metadata.ID]; wasDeleted {
		if ts.After(videoDeleted.Deleted) { // A deleted video will not change its stats, use the latest stats forever.
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
	}
	return VideoStat{}, nil
}

// findClosestTimestamp will resolve the closest available statistic entry at the given ts (timestamp)
// the tolerance of days being between the ts requested and the found ts can be configured either using the Database struct of the StatsIO package or during execution with a flag.
//func findClosestTimestamp(ts time.Time, id int64) (closestStat VideoStat, err error) { // TODO: REIMPLEMENT!
//	var ok bool
//	// Look in the past and in the future relative to the requested timestamp for a replacement of the missing data.
//	for i := 1; i < Database.StatsMissTolerance || !ok; i++ {
//		// left
//		closestStat, ok = getStatOfDate(ts, id)
//		if ok {
//			return
//		}
//		// handle year out of bounds past
//		if targetTs := ts.AddDate(0, 0, -i); targetTs.Before(year.StartDate) {
//			closestStat, ok = years[targetTs.Year()].Stats[statsFormat(targetTs)]
//			if ok {
//				return
//			}
//		}
//		// right
//		closestStat, ok = year.Stats[statsFormat(ts.AddDate(0, 0, i))]
//		if ok {
//			return closestStat, nil
//		}
//		if targetTs := ts.AddDate(0, 0, -i); year.EndDate.Before(targetTs) {
//			closestStat, ok = years[targetTs.Year()].Stats[statsFormat(targetTs)]
//		}
//	}
//
//	return closestStat, nil
//}

func getStatOfDate(ts time.Time, id int64) (result VideoStat, found bool) {
	if _, err := os.Stat(Database.getRawFilePath(ts)); !os.IsNotExist(err) {
		videos := Database.ReadRawResponses(ts)
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
