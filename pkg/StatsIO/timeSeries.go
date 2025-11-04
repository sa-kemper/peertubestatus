package StatsIO

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/sa-kemper/peertubestats/internal/LogHelp"
)

type LikeView struct {
	Likes int64 `json:"likes"`
	Views int64 `json:"views"`
}

func (lv *LikeView) Equal(second *LikeView) bool {
	if second.Likes != lv.Likes {
		return false
	}
	if second.Views != lv.Views {
		return false
	}
	return true
}

type TimeSeriesDataEntry struct {
	Date time.Time            `json:"date"`
	Data LikeView             `json:"data"`
	Next *TimeSeriesDataEntry `json:"next"`
	Prev *TimeSeriesDataEntry `json:"prev"`
}

type DoubleLinkedList struct {
	Head     *TimeSeriesDataEntry `json:"head"`
	Tail     *TimeSeriesDataEntry `json:"tail"`
	Earliest time.Time            `json:"earliest"`
	Latest   time.Time            `json:"latest"`
}

type TimeSeriesDatabase struct {
	// Video is a map[int64]*DoubleLinkedList
	Video          *sync.Map
	FirstTimestamp time.Time
	LastTimestamp  time.Time
}

const TimeSeriesDatabaseFileName = "TimeSeriesDB.json"

// appendHeadTimeSeries inserts an item into the front of the Double linked list IF:
// - The provided timestamp is newer then the earliest in the Double linked list
// --- If the data is a duplicate the earliest is replaced with the provided value
func appendHeadTimeSeries(list *DoubleLinkedList, timestamp time.Time, value *TimeSeriesDataEntry) {
	if list == nil || timestamp.IsZero() || value == nil {
		return
	}
	if !timestamp.Before(list.Earliest) {
		return
	}
	if list.Head.Data.Equal(&value.Data) {
		// If the data is a duplicate only allow the earliest data sample.
		list.Head.Next.Prev = value
		value.Next = list.Head
		list.Head = value
		list.Earliest = timestamp
		return
	}

	list.Head.Prev = value
	value.Next = list.Head
	list.Head = value
	list.Earliest = timestamp
}

// appendTailTimeSeries inserts an item into the end of the Double linked list IF:
// - The data is not a duplicate
// - The data is newer then the last entry ( Tail of the Double linked list )
func appendTailTimeSeries(list *DoubleLinkedList, timestamp time.Time, value *TimeSeriesDataEntry) {
	if list == nil || timestamp.IsZero() || value == nil {
		return
	}
	if !list.Tail.Date.Before(timestamp) {
		return
	}
	if list.Tail.Data.Equal(&value.Data) {
		return
	}

	list.Tail.Next = value
	value.Prev = list.Tail

	list.Tail = value
	list.Latest = value.Date
}

// insertTimeSeries
func insertTimeSeries(list *DoubleLinkedList, timestamp time.Time, value *TimeSeriesDataEntry) {
	if list == nil || timestamp.IsZero() || value == nil {
		LogHelp.NewLog(LogHelp.Warn, "invalid input for insertTimeSeries", map[string]interface{}{"list: ": list, "time": timestamp, "value": value}).Log()
		return
	}
	if list.Tail == nil || list.Head == nil || list.Earliest.IsZero() || list.Latest.IsZero() {
		// We are the first entry.
		list.Head = value
		list.Tail = value
		list.Earliest = timestamp
		list.Latest = timestamp
		return
	}

	if timestamp.Before(list.Earliest) {
		appendHeadTimeSeries(list, timestamp, value)
		return
	}
	if timestamp.After(list.Latest) {
		appendTailTimeSeries(list, timestamp, value)
		return
	}
	var current, previous *TimeSeriesDataEntry
	current = list.Head

	for current.Date.Before(timestamp) && current.Next != nil {
		if current.Next.Date.After(timestamp) {
			if !current.Data.Equal(&value.Data) {
				if previous != nil {
					previous.Next = value
				}
				current.Prev = value
				return
			}
			// skip the value as it is a duplicate
		}
		// we have not found the most current data point that could need updating
		current = current.Next
		previous = current
	}

}

func lookupTimeSeriesSingle(list *DoubleLinkedList, timestamp time.Time) LikeView {
	if list == nil || timestamp.Before(list.Earliest) {
		return LikeView{}
	}
	if list.Tail == nil {
		return LikeView{}
	} // there is no likes or views recorded.
	if timestamp.After(list.Latest) {
		return list.Tail.Data
	}

	distanceHead := timestamp.Sub(list.Earliest)
	distanceTail := timestamp.Sub(list.Latest)

	if distanceHead > distanceTail {
		return findFromTail(list, timestamp)
	}
	return findFromHead(list, timestamp)
}

func findFromTail(list *DoubleLinkedList, timestamp time.Time) LikeView {
	current := list.Tail
	for current.Date.After(timestamp) {
		if current.Prev == nil {
			return LikeView{}
		}
		if current.Prev.Date.Before(timestamp) {
			return current.Prev.Data
		}
		current = current.Prev
	}
	if current == list.Head {
		LogHelp.NewLog(LogHelp.Error, "Cannot find suitable datapoint for timestamp", map[string]string{"Timestamp:": timestamp.Format(time.RFC3339)}).Log()
	}
	return LikeView{}
}

func findFromHead(list *DoubleLinkedList, timestamp time.Time) LikeView {
	current := list.Head
	for current.Date.Before(timestamp) {
		if current.Next == nil || current.Next.Date.After(timestamp) {
			return current.Data
		}
		current = current.Next
	}
	if current == list.Tail {
		LogHelp.NewLog(LogHelp.Error, "Cannot find suitable datapoint for timestamp", map[string]string{"Timestamp:": timestamp.Format(time.RFC3339)}).Log()
	}
	return LikeView{}
}

func serializeTimeSeries(list *TimeSeriesDatabase) error {
	var serialData = struct {
		VideosSaved []int64
		FirstItem   time.Time
		LastItem    time.Time
	}{
		VideosSaved: []int64{},
		FirstItem:   list.FirstTimestamp,
		LastItem:    list.LastTimestamp,
	}
	waitGroup := sync.WaitGroup{}
	err := os.MkdirAll(path.Join(Database.DataFolder, "TimeSeries"), 0700)
	if err != nil {
		return err
	}

	sem := make(chan struct{}, Database.StatIOMaxThreads)
	list.Video.Range(func(key, value interface{}) bool {
		waitGroup.Add(1)
		dllVal, ok := value.(*DoubleLinkedList)
		if !ok {
			LogHelp.NewLog(LogHelp.Fatal, "cannot cast time series value", list).Log()
		}
		vidIDVal, ok := key.(int64)
		if !ok {
			LogHelp.NewLog(LogHelp.Fatal, "cannot cast time series value", list).Log()
		}

		serialData.VideosSaved = append(serialData.VideosSaved, vidIDVal)

		sem <- struct{}{}
		go func() {
			LogHelp.LogOnError("cannot serialize double linked list", nil, serializeDoubleLinkedList(vidIDVal, dllVal, &waitGroup))
			<-sem
		}()
		return true
	})

	handle, err := os.OpenFile(path.Join(Database.DataFolder, TimeSeriesDatabaseFileName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer handle.Close()
	err = json.NewEncoder(handle).Encode(serialData)
	if err != nil {
		return err
	}
	waitGroup.Wait()
	return nil
}

type serializedDoubleLinkedListProxyStruct struct {
	Items map[int64]struct {
		Date time.Time `json:"date"`
		Data LikeView  `json:"data"`
	} `json:"items"`
	Earliest time.Time `json:"earliest"`
	Latest   time.Time `json:"last"`
}

func serializeDoubleLinkedList(id int64, list *DoubleLinkedList, group *sync.WaitGroup) error {
	defer group.Done()

	proxy := serializedDoubleLinkedListProxyStruct{
		Earliest: list.Earliest,
		Latest:   list.Latest,
		Items: make(map[int64]struct {
			Date time.Time `json:"date"`
			Data LikeView  `json:"data"`
		}),
	}
	current := list.Head
	var counter int64 = 1
	handle, err := os.OpenFile(path.Join(Database.DataFolder, "TimeSeries", strconv.FormatInt(id, 10))+".json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer handle.Close()
	for current != nil {
		proxy.Items[counter] = struct {
			Date time.Time `json:"date"`
			Data LikeView  `json:"data"`
		}{Date: current.Date, Data: current.Data}
		current = current.Next
		counter++
	}
	err = json.NewEncoder(handle).Encode(proxy)
	if err != nil {
		return err
	}

	return nil
}

func loadDoubleLinkedList(id int64, group *sync.WaitGroup, store *sync.Map) error {
	defer group.Done()
	var list serializedDoubleLinkedListProxyStruct
	handle, err := os.OpenFile(path.Join(Database.DataFolder, "TimeSeries", strconv.FormatInt(id, 10))+".json", os.O_RDONLY, 0600)
	if err != nil {
		return err
	}
	defer handle.Close()

	err = json.NewDecoder(handle).Decode(&list)
	if err != nil {
		return err
	}
	if len(list.Items) == 0 {
		return nil // no views or likes recorded.
	}
	var current = &TimeSeriesDataEntry{
		Date: list.Items[1].Date,
		Data: list.Items[1].Data,
		Next: nil,
		Prev: nil,
	}
	dll := DoubleLinkedList{
		Head:     current,
		Tail:     current,
		Earliest: list.Earliest,
		Latest:   list.Latest,
	}

	for i := 2; i <= len(list.Items); i++ {
		current.Next = &TimeSeriesDataEntry{
			Date: list.Items[int64(i)].Date,
			Data: list.Items[int64(i)].Data,
			Next: nil,
			Prev: current,
		}
		current = current.Next
		dll.Tail = current
	}
	store.Swap(id, &dll)
	return nil
}

func loadTimeSeries() (*TimeSeriesDatabase, error) {
	var TSDB TimeSeriesDatabase
	var serialData struct {
		VideosSaved []int64
		FistItem    time.Time
		LastItem    time.Time
	}
	waitGroup := sync.WaitGroup{}
	err := os.MkdirAll(path.Join(Database.DataFolder, "TimeSeries"), 0700)
	if err != nil {
		return nil, err
	}

	handle, err := os.OpenFile(path.Join(Database.DataFolder, TimeSeriesDatabaseFileName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		LogHelp.NewLog(LogHelp.Warn, "cannot read time series data, importing from raw", nil).Log()
		return importTimeSeriesFromRawData()
	}
	defer handle.Close()
	err = json.NewDecoder(handle).Decode(&serialData)
	if errors.Is(err, io.EOF) {
		LogHelp.NewLog(LogHelp.Warn, "cannot read time series data, importing from raw", nil).Log()
		return importTimeSeriesFromRawData()
	}
	TSDB.Video = &sync.Map{}
	if err != nil {
		return nil, err
	}
	sem := make(chan struct{}, Database.StatIOMaxThreads)
	waitGroup.Add(len(serialData.VideosSaved))
	for _, id := range serialData.VideosSaved {
		TSDB.Video.Store(id, &DoubleLinkedList{})
		sem <- struct{}{}
		go func() {

			LogHelp.FatalOnError("cannot load double linked list", nil, loadDoubleLinkedList(id, &waitGroup, TSDB.Video))
			<-sem
		}()
	}

	waitGroup.Wait()
	return &TSDB, nil
}

func importTimeSeriesFromRawData() (*TimeSeriesDatabase, error) {
	TsDB := TimeSeriesDatabase{
		Video:          &sync.Map{},
		FirstTimestamp: time.Time{},
		LastTimestamp:  time.Time{},
	}
	var firstDate = Database.firstDataAvailable
	if firstDate.IsZero() {
		firstDate = findFirstDataAvailable()
	}
	var currentDate = firstDate
	for currentDate.Before(time.Now()) {
		Videos := readRawResponses(currentDate)
		if len(Videos) < 1 {
			currentDate = currentDate.AddDate(0, 0, 1)
			continue
		}
		for _, video := range Videos {
			doubleLinkedListVal, found := TsDB.Video.Load(video.ID)
			var doubleLinkedList *DoubleLinkedList
			if found {
				doubleLinkedList = doubleLinkedListVal.(*DoubleLinkedList)
			} else {
				doubleLinkedList = &DoubleLinkedList{}
			}
			insertTimeSeries(doubleLinkedList, currentDate, &TimeSeriesDataEntry{
				Date: currentDate,
				Data: LikeView{
					Likes: video.Likes,
					Views: video.Views,
				},
				Next: nil,
				Prev: nil,
			})
			if !found {
				TsDB.Video.Store(video.ID, doubleLinkedList)
			}
		}
		currentDate = currentDate.AddDate(0, 0, 1)
	}
	err := serializeTimeSeries(&TsDB)
	LogHelp.FatalOnError("cannot serialize time series database", nil, err)
	return &TsDB, nil
}
