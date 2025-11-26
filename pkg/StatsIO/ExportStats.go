package StatsIO

import (
	"cmp"
	"errors"
	"slices"
	"time"
)

func ExportStats(videoID int64, Dates Timeframe, Timeframe string) (Bucket []VideoStat, err error) {
	var dateVals = []int{0, 0, 0}
	if Timeframe != "Daily" && Timeframe != "Monthly" && Timeframe != "Yearly" {
		Timeframe = "Daily"
	}
	switch Timeframe {
	case "Daily":
		dateVals = []int{0, 0, -1}
	case "Monthly":
		dateVals = []int{0, -1, 0}
	case "Yearly":
		dateVals = []int{-1, 0, 0}
	}

	var timestamps = make([]time.Time, 0)
	var currentDate = Dates.GetEndDate()
	var startDate = Dates.GetStartDate()

	if startDate.IsZero() || currentDate.Before(startDate) {
		startDate = time.Now().AddDate(dateVals[0]*4, dateVals[1]*5, dateVals[2]*6)
	}
	if currentDate.IsZero() {
		currentDate = time.Now().AddDate(0, 0, 0)
	}

	for currentDate.After(startDate) || startDate.Equal(currentDate) {
		timestamps = append(timestamps, currentDate)
		currentDate = currentDate.AddDate(dateVals[0], dateVals[1], dateVals[2])
	}

	if len(timestamps) == 0 {
		return make([]VideoStat, 0), errors.New("timestamps is empty")
	}

	slices.SortFunc(timestamps, func(a, b time.Time) int {
		return cmp.Compare(a.Unix(), b.Unix())
	})

	for _, timestamp := range timestamps {
		stat, err := requestTimestamp(timestamp, videoID)
		if err != nil {
			return []VideoStat{}, err
		}
		Bucket = append(Bucket, stat)
	}

	Bucket = prepareStatsForViewing(Bucket)
	return Bucket, nil

}

func prepareStatsForViewing(bucket []VideoStat) []VideoStat {
	var LikesSmallest, LikesBiggest int64
	var ViewsSmallest, ViewsBiggest int64
	for _, stat := range bucket {
		LikesSmallest = min(LikesSmallest, stat.Likes.Data)
		LikesBiggest = max(LikesBiggest, stat.Likes.Data)

		ViewsSmallest = min(ViewsSmallest, stat.Views.Data)
		ViewsBiggest = max(ViewsBiggest, stat.Views.Data)
	}

	for i, stat := range bucket {
		likesCurrentPercent := float64(stat.Likes.Data) / max(float64(1), float64(LikesBiggest))
		viewsCurrentPercent := float64(stat.Views.Data) / max(float64(1), float64(ViewsBiggest))

		if i+1 == len(bucket) {
			bucket[i].Views.EndPercentage = viewsCurrentPercent
			bucket[i].Likes.EndPercentage = likesCurrentPercent

			bucket[i].Views.EndPercentage = viewsCurrentPercent
			bucket[i].Likes.EndPercentage = likesCurrentPercent
			break
		}

		if i == 0 {
			bucket[i].Likes.StartPercentage = likesCurrentPercent
			bucket[i].Views.StartPercentage = viewsCurrentPercent
		}
		bucket[i].Likes.EndPercentage = likesCurrentPercent
		bucket[i].Views.EndPercentage = viewsCurrentPercent

		bucket[i+1].Likes.StartPercentage = likesCurrentPercent
		bucket[i+1].Views.StartPercentage = viewsCurrentPercent
	}
	return bucket
}

func PrepareStatsBucketWithAverages(bucket []VideoStat) []VideoStat {
	return prepareStatsForViewing(bucket)
}
