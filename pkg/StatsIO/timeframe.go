package StatsIO

import "time"

type Timeframe interface {
	GetStartDate() time.Time
	GetEndDate() time.Time
}
