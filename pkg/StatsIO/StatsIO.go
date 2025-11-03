package StatsIO

import (
	"flag"
	"os"
	"path"
	"sync"
	"time"

	"github.com/sa-kemper/peertubestats/pkg/peertubeApi"
)

type VideoDatabase map[int64]peertubeApi.VideoData

func (db *VideoDatabase) VideoIsDeleted(video peertubeApi.VideoData) *bool {
	dletedDB, err := loadDeletedDB()
	if err != nil {
		return nil
	}
	_, deleted := dletedDB[video.ID]
	return &deleted
}

func (db *VideoDatabase) VideoExists(data peertubeApi.VideoData) (exists bool) {
	_, exists = (*db)[data.ID]
	return
}

func (db *VideoDatabase) VideoAdd(data peertubeApi.VideoData) {
	(*db)[data.ID] = data
}

type StatsIO struct {
	DataFolder               string
	StatsMissTolerance       int
	data                     *VideoDatabase
	dataMutex                sync.RWMutex
	CacheInvalidationSeconds int
	Api                      *peertubeApi.ApiClient
	firstDataAvailable       time.Time
}

func (statIO *StatsIO) Init(api *peertubeApi.ApiClient) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		statIO.TimeSeriesDB, err = loadTimeSeries()
		LogHelp.FatalOnError("cannot load time series database", nil, err)

	}()
	db, err := loadVideoDB()
	if err == nil {
		Database.data = db
	}
	Database.firstDataAvailable = findFirstDataAvailable()
	if api != nil {
		statIO.Api = api
	}
	wg.Wait()
}

func (statIO *StatsIO) ReadRawResponsesByPath(p string, i *[]peertubeApi.VideoResponse) (err error) {
	if i == nil {
		return errors.New("invalid input")
	}
	var FileBytes []byte
	FileBytes, err = os.ReadFile(p)
	if err != nil {
		return err
	}
	versionIndex := bytes.IndexByte(FileBytes, byte('\n'))
	if versionIndex == -1 {
		LogHelp.NewLog(LogHelp.Error, "cannot find version header of raw data", map[string]string{"Path": p}).Log()
		versionIndex = 0
	}
	decoder := json.NewDecoder(bytes.NewReader(FileBytes[versionIndex+1:]))
	for {
		var video peertubeApi.VideoResponse
		err = decoder.Decode(&video)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			LogHelp.LogOnError("error parsing imported data", nil, err)
			return
		}
		*i = append(*i, video)
	}
	return
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
