package utils

import "time"

type DateIter struct {
	Start       time.Time
	End         time.Time
	Interval    string
	CurrentTime time.Time
	Index       int32
}

const DAYS = "DAYS"

const MONTHS = "MONTHS"

const YEARS = "YEARS"

func DateRangeNew(start time.Time, end time.Time, interval string) *DateIter {
	return &DateIter{
		Start:       start,
		End:         end,
		Interval:    interval,
		CurrentTime: start,
		Index:       0,
	}
}

func (i *DateIter) Next() bool {
	var next time.Time

	if i.Index == 0 {
		next = i.CurrentTime
	} else {
		if i.Interval == DAYS {
			next = i.CurrentTime.AddDate(0, 0, 1)
		} else if i.Interval == MONTHS {
			next = i.CurrentTime.AddDate(0, 1, 0)
		} else if i.Interval == YEARS {
			next = i.CurrentTime.AddDate(0, 1, 0)
		}
	}

	if i.End.Equal(next) || i.End.After(next) {
		i.CurrentTime = next
		i.Index++
		return true
	}

	return false
}

func (i *DateIter) Current() time.Time {
	return i.CurrentTime
}
