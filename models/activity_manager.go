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
	Activities(since, typeOf string) []data.Activity
	AppendActivityRow(activity *data.Activity) error
	UpdateActivityStartTime(activity data.Activity)
	UpdateActivityStopTime(activity data.Activity)
	FirstActivity(since, typeOf string) *data.Activity
	LastActivity(typeOf string) *data.Activity
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
	lastBreak := m.LastBreak()
	if lastBreak == nil {
		log.Panic("Can't start the break. Set break is missing.")
	}
	lastBreak.Start = time.Now()

	m.db.UpdateActivityStartTime(*lastBreak)
}

func (m ActivityManager) SetBreak() {
	activity := &data.Activity{
		Type: data.ACTIVITY_TYPE_BREAK,
	}

	log.Println("Set break")
	m.addActivity(activity)
}

func (m ActivityManager) FinishActivity(activity *data.Activity) {
	if activity.Start.IsZero() || activity.Type == "" {
		log.Panic("Invalid activity to finish")
	}
	activity.Stop = time.Now()

	m.db.UpdateActivityStopTime(*activity)
}

func (m ActivityManager) FirstActivity(typeOf data.ActivityType) *data.Activity {
	now := FormattedDatetime(time.Now())
	return m.db.FirstActivity(now, string(typeOf))
}

func (m ActivityManager) lastActivity(typeOf data.ActivityType) *data.Activity {
	return m.db.LastActivity(string(typeOf))
}

func (m ActivityManager) LastActivity() *data.Activity {
	return m.lastActivity(data.ACTIVITY_TYPE_ANY)
}

func (m ActivityManager) FirstSession() *data.Activity {
	return m.FirstActivity(data.ACTIVITY_TYPE_SESSION)
}

func (m ActivityManager) LastSession() *data.Activity {
	return m.lastActivity(data.ACTIVITY_TYPE_SESSION)
}

func (m ActivityManager) FirstBreak() *data.Activity {
	return m.FirstActivity(data.ACTIVITY_TYPE_BREAK)
}

func (m ActivityManager) LastBreak() *data.Activity {
	return m.lastActivity(data.ACTIVITY_TYPE_BREAK)
}

func (m ActivityManager) LeaveTime() time.Time {
	firstActivity := m.FirstSession()
	return firstActivity.Start.Add(WORKING_HOURS_LENGTH*time.Hour + m.BreakTime())
}

func (m ActivityManager) WorkTime() time.Duration {
	firstSessionStart := m.FirstSession().Start
	now := time.Now()

	//TODO
	total := time.Duration(now.Hour())*time.Hour +
		time.Duration(now.Minute())*time.Minute +
		time.Duration(now.Second())*time.Second -
		(time.Duration(firstSessionStart.Hour())*time.Hour +
			time.Duration(firstSessionStart.Minute())*time.Minute +
			time.Duration(firstSessionStart.Second())*time.Second)

	return total - m.BreakTime()
}

func (m ActivityManager) BreakTime() time.Duration {
	var duration time.Duration

	breaks := m.db.Activities(FormattedDatetime(time.Now()), data.ACTIVITY_TYPE_BREAK)
	for _, b := range breaks {
		duration += b.Stop.Sub(b.Start)
	}

	log.Println("Break time: ", duration)

	return duration
}

func (m ActivityManager) Activities(date time.Time) []data.Activity {
	return m.db.Activities(FormattedDatetime(date), data.ACTIVITY_TYPE_ANY)
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

	return firstIdx
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

func DailyOvertime(activities []data.Activity, date time.Time) time.Duration {

	for _, v := range activities {
		log.Println(v)
	}
	log.Println("Get overtime from: ", date.String())

	firstSessionIdx := firstActivityOfTheDayIdx(activities, data.ACTIVITY_TYPE_SESSION, date)
	if firstSessionIdx < 0 {
		log.Println("First activity not found")
		return time.Duration(0)
	}

	lastSessionIdx := firstSessionIdx + lastActivityOfTheDayIdx(activities[firstSessionIdx:], data.ACTIVITY_TYPE_SESSION, date)
	if lastSessionIdx < 0 || firstSessionIdx > lastSessionIdx {
		log.Println("Last activity not found")
		return time.Duration(0)
	}

	log.Println("first: ", firstSessionIdx, " last: ", lastSessionIdx)

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

	firstDayOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	activities := m.db.Activities(FormattedDatetime(firstDayOfMonth), string(data.ACTIVITY_TYPE_ANY))

	day := firstDayOfMonth

	for day.Before(now) {

		ot := DailyOvertime(activities, day)
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
