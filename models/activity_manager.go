package models

import (
	"log"
	"time"

	"github.com/kamilez/day_watch/data"
	. "github.com/kamilez/day_watch/utils"
)

const (
	WORKING_HOURS_LENGTH = 8
	WEEK_DAYS_COUNT      = 7
)

type ActivityProvider interface {
	Activities(since, till, typeOf string) []data.Activity
	AppendActivityRow(activity *data.Activity) error
	UpdateActivityStartTime(activity data.Activity)
	UpdateActivityStopTime(activity data.Activity)
}

type ActivityManager struct {
	db ActivityProvider
}

func NewActivityManager(db ActivityProvider) *ActivityManager {

	return &ActivityManager{db: db}
}

func (m ActivityManager) addActivity(activity *data.Activity) {
	m.db.AppendActivityRow(activity)
}

func (m ActivityManager) StartSession() {

	activity := &data.Activity{
		Start: time.Now(),
		Type:  data.ACTIVITY_TYPE_SESSION,
	}

	m.addActivity(activity)
}

func (m ActivityManager) StartBreak() {

	breaks := m.dayActivities(time.Now(), data.ACTIVITY_TYPE_BREAK)
	length := len(breaks)
	if length == 0 {
		log.Panic("No breaks set today")
	}

	lastBreak := breaks[length-1]
	if lastBreak.Start.IsZero() == true && lastBreak.Stop.IsZero() == true {
		lastBreak.Start = time.Now()
		m.db.UpdateActivityStartTime(lastBreak)
	} else {
		log.Panic("Last break is already started or finished. Set new break to start.")
	}
}

func (m ActivityManager) SetBreak() {
	activity := &data.Activity{
		Type: data.ACTIVITY_TYPE_BREAK,
	}

	m.addActivity(activity)
}

func (m ActivityManager) FinishActivity(activity *data.Activity) {
	if activity.Start.IsZero() || activity.Type == "" {
		log.Panic("Invalid activity to finish")
	}
	activity.Stop = time.Now()

	m.db.UpdateActivityStopTime(*activity)
}

func startOfDay(t time.Time) time.Time {
	return t.Truncate(24 * time.Hour)
}

func endOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.Local)
}

func (m ActivityManager) dayActivities(date time.Time, aType data.ActivityType) []data.Activity {

	activities := m.db.Activities(
		FormattedDatetime(startOfDay(date)),
		FormattedDatetime(endOfDay(date)),
		string(aType))

	if len(activities) == 0 {
		log.Println("No activities since ", startOfDay(date), " and till ", endOfDay(date))
		return []data.Activity{}
	}

	return activities
}

func (m ActivityManager) FirstActivity(aType data.ActivityType) *data.Activity {

	activities := m.dayActivities(time.Now(), aType)
	if len(activities) == 0 {
		return nil
	}

	return &activities[0]
}

func (m ActivityManager) LastActivity(aType data.ActivityType) *data.Activity {
	activities := m.dayActivities(time.Now(), aType)

	length := len(activities)
	if length == 0 {
		return nil
	}

	return &activities[length-1]
}

func (m ActivityManager) LeaveTime() time.Time {
	firstSession := m.FirstActivity(data.ACTIVITY_TYPE_SESSION)
	if firstSession == nil {
		log.Panic("Can't find first session of the day")
	}

	return firstSession.Start.Add(WORKING_HOURS_LENGTH*time.Hour + m.BreakTime())
}

func (m ActivityManager) WorkTime() time.Duration {
	firstSession := m.FirstActivity(data.ACTIVITY_TYPE_SESSION)
	workTime := time.Now().Sub(firstSession.Start)

	return workTime - m.BreakTime()
}

func (m ActivityManager) BreakTime() (duration time.Duration) {

	breaks := m.dayActivities(time.Now(), data.ACTIVITY_TYPE_BREAK)
	for _, b := range breaks {
		duration += b.Stop.Sub(b.Start)
	}

	return
}

func (m ActivityManager) Activities(date time.Time) []data.Activity {
	return m.dayActivities(date, data.ACTIVITY_TYPE_ANY)
}

func firstActivityOfTheDayIdx(activities []data.Activity, aType data.ActivityType, date time.Time) int {

	if len(activities) == 0 {
		return -1
	}

	//Assuming activities are in chronological order
	for k, a := range activities {
		if a.Start.Year() == date.Year() && a.Start.YearDay() == date.YearDay() {
			return k
		}
	}

	return -1
}

func lastActivityOfTheDayIdx(activities []data.Activity, aType data.ActivityType, date time.Time) int {

	if len(activities) == 0 {
		return -1
	}

	//Assuming activities are in chronological order
	firstIdx := firstActivityOfTheDayIdx(activities, aType, date)
	if firstIdx < 0 {
		return -1
	}

	lastIdx := firstIdx
	for k, a := range activities[firstIdx:] {
		if a.Start.Year() != date.Year() || a.Start.YearDay() != date.YearDay() {
			return lastIdx
		}
		lastIdx = k
	}

	return lastIdx
}

func breakTime(activities []data.Activity) time.Duration {

	var duration time.Duration
	for _, a := range activities {
		if a.Type == data.ACTIVITY_TYPE_BREAK {
			if a.Stop.IsZero() {
				continue
			}
			duration += a.Stop.Sub(a.Start)
		}
	}

	return duration
}

func dayOvertime(activities []data.Activity, date time.Time) time.Duration {

	firstSessionIdx := firstActivityOfTheDayIdx(activities, data.ACTIVITY_TYPE_SESSION, date)
	if firstSessionIdx < 0 {
		return time.Duration(0)
	}

	lastSessionIdx := firstSessionIdx + lastActivityOfTheDayIdx(activities[firstSessionIdx:], data.ACTIVITY_TYPE_SESSION, date)
	if lastSessionIdx < 0 || firstSessionIdx > lastSessionIdx {
		return time.Duration(0)
	}

	lastSession := activities[lastSessionIdx]
	if lastSession.Stop.IsZero() {
		return time.Duration(0)
	}

	workTime := lastSession.Stop.Sub(activities[firstSessionIdx].Start)
	breakTime := breakTime(activities[firstSessionIdx : lastSessionIdx+1])

	workTime -= breakTime

	weekday := date.Weekday()
	if weekday >= time.Monday && weekday <= time.Friday {
		return workTime - WORKING_HOURS_LENGTH*time.Hour
	} else if weekday == time.Saturday {
		if workTime >= 4*time.Hour {
			return WORKING_HOURS_LENGTH * time.Hour
		} else {
			return workTime
		}
	} else {
		return workTime
	}
}

func (m ActivityManager) Overtime() (overtime, saturdayOvertime, sundayOvertime time.Duration) {
	now := time.Now()

	firstDayOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	activities := m.db.Activities(FormattedDatetime(firstDayOfMonth), FormattedDatetime(endOfDay(now.AddDate(0, 0, -1))), string(data.ACTIVITY_TYPE_ANY))

	day := firstDayOfMonth

	for day.Before(now) {

		ot := dayOvertime(activities, day)
		if ot != time.Duration(0) {
			weekday := day.Weekday()
			if weekday >= time.Monday && weekday <= time.Friday {
				overtime += ot
			} else if weekday == time.Saturday {
				saturdayOvertime += ot
			} else {
				sundayOvertime += ot
			}
		}

		day = day.AddDate(0, 0, 1)
	}

	return
}
