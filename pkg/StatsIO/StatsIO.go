package StatsIO

import (
	"flag"
	"os"
	"path"
	"sync"
	"time"

	"github.com/sa-kemper/peertubestats/pkg/peertubeapi"
)

type VideoDatabase map[int64]peertubeapi.VideoData

func (db *VideoDatabase) VideoIsDeleted(video peertubeapi.VideoData) *bool {
	dletedDB, err := loadDeletedDB()
	if err != nil {
		return nil
	}
	_, deleted := dletedDB[video.ID]
	return &deleted
}

func (db *VideoDatabase) VideoExists(data peertubeapi.VideoData) (exists bool) {
	_, exists = (*db)[data.ID]
	return
}

func (db *VideoDatabase) VideoAdd(data peertubeapi.VideoData) {
	(*db)[data.ID] = data
}

type StatsIO struct {
	DataFolder               string
	StatsMissTolerance       int
	data                     *VideoDatabase
	dataMutex                sync.RWMutex
	CacheInvalidationSeconds int
	Api                      *peertubeapi.ApiClient
	firstDataAvailable       time.Time
}

func (statIO *StatsIO) Init() {
	db, err := loadVideoDB()
	if err == nil {
		Database.data = &db
	}
	Database.firstDataAvailable = findFirstDataAvailable()
}

var Database StatsIO

func init() {
	flag.StringVar(&Database.DataFolder, "data-folder", "./Data", "Folder containing video stats")
	flag.IntVar(&Database.StatsMissTolerance, "miss-tolerance", 0, "If a searched statistic is missing, this specifies the tolerance of days of a mismatch before an error.")
	flag.IntVar(&Database.CacheInvalidationSeconds, "cache-valid-seconds", 1*60*60*25, "The number of seconds the video database cache is valid, By default a bit more than a day")
}

func findFirstDataAvailable() time.Time {
	currentDate := time.Now()
	// Find the oldest year
	for {
		if _, err := os.Stat(path.Join(Database.DataFolder, currentDate.Format("2006"))); os.IsNotExist(err) {
			break
		}
		currentDate = currentDate.AddDate(-1, 0, 0)
	}
	currentDate = currentDate.AddDate(1, 0, 0)

	// find the oldest month
	for {
		if _, err := os.Stat(path.Join(Database.DataFolder, currentDate.Format("2006"), currentDate.Format("01"))); os.IsNotExist(err) {
			break
		}
		currentDate = currentDate.AddDate(0, -1, 0)
	}
	currentDate = currentDate.AddDate(0, 1, 0)

	// find the oldest day
	for {
		if _, err := os.Stat(path.Join(Database.DataFolder, currentDate.Format("2006"), currentDate.Format("01"))); os.IsNotExist(err) {
			break
		}
		currentDate = currentDate.AddDate(0, 0, -1)
	}
	currentDate = currentDate.AddDate(0, 0, 1)
	return currentDate

}
