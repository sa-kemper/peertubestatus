package StatsIO

import (
	"errors"
	"time"
)

func collectTimestampsForID(id int64, timestamps []time.Time) (Stats []VideoStat, err error) {
	if len(timestamps) == 0 {
		return make([]VideoStat, 0), errors.New("timestamps is empty")
	}

	for _, timestamp := range timestamps {
		stat, err := requestTimestamp(timestamp, id)
		if err != nil {
			return []VideoStat{}, err
		}
		Stats = append(Stats, stat)
	}

	return Stats, nil
}
