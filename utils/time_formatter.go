package utils

import (
	"log"
	"time"
)

const (
	DATE_FORMAT             = "2006-01-02"
	TIME_FORMAT             = "15:04:05"
	SQLITE3_DATETIME_FORMAT = "2006-01-02 15:04:05"
)

func FormattedTime(time time.Time) string {
	if time.IsZero() == true {
		return ""
	}
	return time.Format(TIME_FORMAT)
}

func FormattedDate(time time.Time) string {
	if time.IsZero() == true {
		return ""
	}
	return time.Format(DATE_FORMAT)
}

func FormattedDatetime(time time.Time) string {
	if time.IsZero() == true {
		return ""
	}
	return time.Format(SQLITE3_DATETIME_FORMAT)
}

func string2DateTime(str, format string) *time.Time {
	if str == "" {
		return &time.Time{}
	}

	t, err := time.Parse(format, str)
	if err != nil {
		log.Panic(err.Error())
	}

	return &t
}

func String2Time(str string) *time.Time {
	return string2DateTime(str, SQLITE3_DATETIME_FORMAT)
}
