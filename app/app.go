package app

import (
	"fmt"
	"log"
	"time"

	"github.com/kamilez/day_watch/data"
	db "github.com/kamilez/day_watch/database"
	. "github.com/kamilez/day_watch/utils"
)

type App struct {
	db       *db.Database
	notifier Notifier
}

func NewApp(db *db.Database, notifier Notifier) *App {
	db.TableCreate("activities", data.Activity{})

	return &App{db, notifier}
}

func (a App) HandleNotification() {

	now := time.Now()
	nowDate := now.Format(DEFAULT_DATE_FORMAT)

	breaks, err := a.db.BreakHours(nowDate)
	if err != nil {
		log.Panic(err.Error())
	}

	firstActivity, err := a.db.FirstActivity(nowDate, ACTIVITY_TYPE_SESSION)
	if err != nil {
		log.Panic(err.Error())
	}

	firstActivityStartTime, err := time.Parse(DEFAULT_TIME_FORMAT, firstActivity.Start)
	if err != nil {
		log.Panic(err.Error())
	}

	leaveTime := firstActivityStartTime.Add(WORKING_HOURS_LENGTH*time.Hour + breaks)
	total, err := a.db.SessionHours(nowDate)
	if err != nil {
		log.Panic(err.Error())
	}

	lastActivity, err := a.db.LastActivity(nowDate, ACTIVITY_TYPE_SESSION)
	if err != nil {
		log.Panic(err.Error())
	}

	lastActivityTime, err := time.Parse(DEFAULT_TIME_FORMAT, lastActivity.Start)
	if err != nil {
		log.Panic(err.Error())
	}

	total += time.Duration(now.Hour()-lastActivityTime.Hour())*time.Hour +
		time.Duration(now.Minute()-lastActivityTime.Minute())*time.Minute

	label := []string{
		"Started:\t" + firstActivityStartTime.Format("15:04"),
		"Leave:\t\t" + leaveTime.Format("15:04"),
		"Work time:\t" + fmt.Sprintf("%02d:%02d",
			int(total.Hours()), int(total.Minutes())%60),
	}

	a.notifier.Notify(label...)
}

func (a App) onTomatoTick(timeLeft time.Duration, type_ SessionType) {

	timeInMinutes := int(timeLeft.Minutes())
	formattedString := fmt.Sprintf("Time left: %d minutes", timeInMinutes)

	if timeInMinutes < 3 || timeInMinutes%5 == 0 {

		var typeName string

		if type_ == WORK {
			typeName = "Work session"
		} else {
			typeName = "Break session"
		}

		a.notifier.Notify(typeName, formattedString)
	}
}

func (a App) HandleTomato() {

	var session *TomatoSession
	for {
		if session == nil || session.Type == BREAK {
			if Dial("Start work session") != nil {
				return
			}
			session = NewTomatoSession(WORK, a.onTomatoTick)
		} else {
			if Dial("Start break session") != nil {
				return
			}
			session = NewTomatoSession(BREAK, a.onTomatoTick)
		}

		<-session.Run()
	}
}

func (a App) HandleLogin() {

	now := time.Now()

	nowDateString := now.Format(DEFAULT_DATE_FORMAT)
	activity := data.Activity{
		Start: now.Format(DEFAULT_TIME_FORMAT),
		Stop:  "",
		Date:  nowDateString,
		Type:  ACTIVITY_TYPE_SESSION,
	}

	lastBreak, err := a.db.LastActivity(nowDateString, ACTIVITY_TYPE_BREAK)
	if err != nil {
		log.Fatalln(err.Error())
	}

	if lastBreak.Stop == "" {
		lastBreak.Stop = now.Format(DEFAULT_TIME_FORMAT)
		a.db.UpdateActivityStopTime(lastBreak)
	}

	a.db.RowAppend("activities", activity)

	a.HandleNotification()
}

func (a App) HandleLogout() {

	now := time.Now()
	nowDateString := now.Format(DEFAULT_DATE_FORMAT)
	nowTimeString := now.Format(DEFAULT_TIME_FORMAT)

	lastActivity, err := a.db.LastActivity(nowDateString, ACTIVITY_TYPE_SESSION)
	if err != nil {
		log.Panic(err.Error())
	}

	if lastActivity.Stop == "" {
		lastActivity.Stop = nowTimeString
		a.db.UpdateActivityStopTime(lastActivity)
	}

	lastActivity, err = a.db.LastActivity(nowDateString, ACTIVITY_TYPE_BREAK)
	if err != nil {
		log.Panic(err.Error())
	}

	if lastActivity.Start == "" {
		lastActivity.Start = nowTimeString
		a.db.UpdateActivityStartTime(lastActivity)
	}
}

func (a App) HandleBreak() {

	now := time.Now()

	nowDateString := now.Format(DEFAULT_DATE_FORMAT)
	activity := data.Activity{
		Start: "",
		Stop:  "",
		Date:  nowDateString,
		Type:  ACTIVITY_TYPE_BREAK,
	}

	activites, err := a.db.Activities(nowDateString)
	if err != nil {
		log.Fatalln(err.Error())
	}

	if activites[len(activites)-1].Type == ACTIVITY_TYPE_BREAK {
		return
	}

	a.db.RowAppend("activities", activity)
}

func (a App) HandleStatus() {

	now := time.Now()
	nowDateString := now.Format(DEFAULT_DATE_FORMAT)

	activities, err := a.db.Activities(nowDateString)
	if err != nil {
		log.Panic(err.Error())
	}

	if len(activities) == 0 {
		activities, err =
			a.db.Activities(now.AddDate(0, 0, -1).Format(DEFAULT_DATE_FORMAT))
		if err != nil {
			log.Panic(err.Error())
		}
	}

	if len(activities) == 0 {
		fmt.Println("Empty data")
		return
	}

	fmt.Println("ID\tTYPE\t\tSTART\t\tSTOP\t\tDATE")
	for k, v := range activities {
		fmt.Printf("%d\t%s\t\t%s\t%s\t%s\n",
			k, v.Type, v.Start, v.Stop, v.Date)
	}

}
