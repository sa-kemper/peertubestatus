package templates

import "time"

type TwoDateForm struct {
	StartDate time.Time `form:"start_date" json:"start_date"`
	EndDate   time.Time `form:"end_date" json:"end_date"`
}

func (t TwoDateForm) GetStartDate() time.Time {
	return t.StartDate
}

func (t TwoDateForm) GetEndDate() time.Time {
	return t.EndDate
}
