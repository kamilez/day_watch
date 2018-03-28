package data

import (
	"time"

	. "github.com/kamilez/day_watch/utils"
)

type ActivityType string

const (
	ACTIVITY_TYPE_SESSION ActivityType = "session"
	ACTIVITY_TYPE_BREAK                = "break"
	ACTIVITY_TYPE_ANY                  = ""
)

type Activity struct {
	Start time.Time
	Stop  time.Time
	Type  ActivityType
}

func (a Activity) StartString() string {
	if a.Start.IsZero() {
		return ""
	}

	return FormattedTime(a.Start)
}

func (a Activity) StopString() string {
	if a.Stop.IsZero() {
		return ""
	}
	return FormattedTime(a.Stop)
}

func (a Activity) DateString() string {
	if a.Start.IsZero() {
		return ""
	}
	return FormattedDate(a.Start)
}

func (a Activity) IsBreak() bool {
	return a.Type == ACTIVITY_TYPE_BREAK
}

func (a Activity) IsSession() bool {
	return a.Type == ACTIVITY_TYPE_SESSION
}

func (a Activity) Weekday() string {
	return a.Start.Weekday().String()
}
