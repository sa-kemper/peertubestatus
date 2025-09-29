package peertubeApi

import (
	"strings"
	"sync"
	"time"
)

type endpointPath string
type RateLimitMap map[endpointPath]*RateLimit

func (rm *RateLimitMap) Match(path endpointPath) *RateLimit {
	var matchLen int
	var match endpointPath
	for key := range *rm {
		if key == path {
			return (*rm)[key]
		}
		if len(key) > matchLen {
			if strings.Contains(string(path), string(key)) && len(key) > matchLen {
				matchLen = len(key)
				match = key
			}
		}
	}
	if matchLen > 0 {
		// We found a somewhat matching path
		return (*rm)[match]
	}

	return nil
}

// RateLimit implements the mentioned https://docs.joinpeertube.org/api-rest-reference.html#section/Rate-limits
type RateLimit struct {
	Requests         int           `json:"requests"`
	TimeFrame        time.Duration `json:"time_frame"`
	Endpoint         endpointPath  `json:"endpoint"`
	mu               sync.Mutex
	requestTimestamp time.Time
	requestsMade     int
}

// NewRateLimit Initializes a RateLimit struct alongside the respective timer, that resets the limit.
func NewRateLimit(Ep endpointPath, Reqs int, Tf time.Duration) (limit *RateLimit) {
	limit = &RateLimit{
		Requests:  Reqs,
		TimeFrame: Tf,
		Endpoint:  Ep,
		mu:        sync.Mutex{},
	}

	go func() {
		ResetTicker := time.NewTicker(Tf)
		defer ResetTicker.Stop()

		for {
			<-ResetTicker.C
			limit.mu.Lock()
			limit.requestsMade = 0
			limit.mu.Unlock()
		}
	}()

	return
}

func (rl *RateLimit) Request() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if rl.requestsMade == 0 {
		rl.requestTimestamp = time.Now()
	}
	rl.requestsMade++

	if rl.requestsMade >= rl.Requests {
		// sleep until the timer ticks
		now := time.Now()
		firstReq := rl.requestTimestamp
		resetTime := firstReq.Add(rl.TimeFrame)
		durationUntilReset := resetTime.Sub(now)
		if !now.After(resetTime) {
			time.Sleep(durationUntilReset)
		}
	}
}
