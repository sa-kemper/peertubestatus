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
	// now never contains the time as we do not care about it, only the date.
	now := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC)

	// If both dates are zero, set them to now
	if fpr.Dates.StartDate.IsZero() && fpr.Dates.EndDate.IsZero() {
		fpr.Dates.StartDate = now
		fpr.Dates.EndDate = now
	}

	switch fpr.Timeframe {
	case "":
		fallthrough
	case "Daily":
		// if no start date was specified, or the timeframe matches a previous calculation, reset it.
		if fpr.Dates.StartDate.Equal(now) || fpr.Dates.StartDate.Equal(now.AddDate(0, -5, 0)) || fpr.Dates.StartDate.Equal(now.AddDate(-4, 0, 0)) {
			fpr.Dates.StartDate = now.AddDate(0, 0, -6)
			fpr.Dates.EndDate = now
		}
		// in case of a fallthrough assume daily
		fpr.Timeframe = "Daily"

	case "Monthly":
		// Reset to 5 months ago if start date is now or date range is too short
		if fpr.Dates.StartDate.Equal(now) || fpr.Dates.EndDate.Sub(fpr.Dates.StartDate).Hours() < ((24*30)*2) || fpr.Dates.StartDate.Equal(now.AddDate(-4, 0, 0)) {
			fpr.Dates.StartDate = now.AddDate(0, -5, 0)
			fpr.Dates.EndDate = now
		}

	case "Yearly":
		// Reset to 4 years ago if start date is now or date range is too short
		if fpr.Dates.StartDate.Equal(now) || fpr.Dates.EndDate.Sub(fpr.Dates.StartDate).Hours() < ((24*30)*12)*2 {
			fpr.Dates.StartDate = now.AddDate(-4, 0, 0)
			fpr.Dates.EndDate = now
		}
	}
}
