package templates

import "time"

type FrontPageRequest struct {
	// Timeframe can be Daily, Monthly or Yearly
	Timeframe string `form:"timeframe" json:"timeframe"`
	// Query the content of the search field
	Query string      `form:"query" json:"query"`
	Dates TwoDateForm `json:"dates" form:"dates"`
}

func (fpr *FrontPageRequest) HandleZeroDate() {
	now := time.Now()
	if fpr.Dates.StartDate.IsZero() {
		fpr.Dates.StartDate = now
	}
	if fpr.Dates.EndDate.IsZero() {
		fpr.Dates.EndDate = now
	}

	switch fpr.Timeframe {
	case "":
		fallthrough
	case "Daily":
		if fpr.Dates.StartDate.Equal(now) {
			fpr.Dates.StartDate = now.AddDate(0, 0, -6)
		}
		// in case of a fallthrough assume daily
		fpr.Timeframe = "Daily"
	case "Monthly":
		if fpr.Dates.StartDate.Equal(now) {
			fpr.Dates.StartDate = now.AddDate(0, -5, 0)
		}
	case "Yearly":
		if fpr.Dates.StartDate.Equal(now) {
			fpr.Dates.StartDate = now.AddDate(-4, 0, 0)
		}

	}
}
