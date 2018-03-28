package app

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/kamilez/day_watch/models"
	. "github.com/kamilez/day_watch/utils"
)

type App struct {
	am       *models.ActivityManager
	notifier Notifier
}

func NewApp(am *models.ActivityManager, notifier Notifier) *App {
	return &App{am, notifier}
}

func (a App) OnError() {
	if err := recover(); err != nil {
		a.notifier.Error(err)
	}
}

func (a App) HandleNotification() {

	defer a.OnError()

	workTime := a.am.WorkTime()
	overtime, saturdayOvertime, sundayOvertime := a.am.Overtime()

	label := []string{
		"Started:\t\t" + FormattedTime(a.am.FirstSession().Start),
		"Leave:\t\t" + FormattedTime(a.am.LeaveTime()),
		"Work time:\t" + fmt.Sprintf("%02d:%02d", int(workTime.Hours()), int(workTime.Minutes())%60),
	}
	if overtime != time.Duration(0) {
		label = append(label, "Overtime:\t"+fmt.Sprintf("%02d:%02d", int(overtime.Hours()), int(math.Abs(overtime.Minutes()))%60))
	}

	if saturdayOvertime != time.Duration(0) {
		label = append(label, "Saturday:\t"+fmt.Sprintf("%02d:%02d", int(saturdayOvertime.Hours()), int(math.Abs(saturdayOvertime.Minutes()))%60))
	}

	if sundayOvertime != time.Duration(0) {
		label = append(label, "Sunday:\t"+fmt.Sprintf("%02d:%02d", int(sundayOvertime.Hours()), int(math.Abs(sundayOvertime.Minutes()))%60))
	}

	a.notifier.Notify(label...)
}

func (a App) onTomatoTick(timeLeft time.Duration, type_ models.SessionType) {

	timeInMinutes := int(timeLeft.Minutes())
	formattedString := fmt.Sprintf("Time left: %d minutes", timeInMinutes)

	if timeInMinutes < 3 || timeInMinutes%5 == 0 {

		var typeName string

		if type_ == models.WORK {
			typeName = "Work session"
		} else {
			typeName = "Break session"
		}

		a.notifier.Notify(typeName, formattedString)
	}
}

func (a App) HandleTomato() {

	defer a.OnError()

	var session *models.TomatoSession
	for {
		if session == nil || session.Type == models.BREAK {
			if Dial("Start work session") != nil {
				return
			}
			session = models.NewTomatoSession(models.WORK, a.onTomatoTick)
		} else {
			if Dial("Start break session") != nil {
				return
			}
			session = models.NewTomatoSession(models.BREAK, a.onTomatoTick)
		}

		<-session.Run()
	}
}

func (a App) HandleLogin() {

	defer a.OnError()

	lastActivity := a.am.LastActivity()
	if lastActivity != nil && lastActivity.IsBreak() == true {
		a.am.FinishActivity(lastActivity)
	}

	a.am.StartSession()
	a.HandleNotification()
}

func (a App) HandleLogout() {

	defer a.OnError()

	lastActivity := a.am.LastActivity()
	if lastActivity != nil && lastActivity.IsBreak() == true {
		a.am.StartBreak()
	}

	lastSession := a.am.LastSession()
	a.am.FinishActivity(lastSession)
}

func (a App) HandleBreak() {

	defer a.OnError()

	lastActivity := a.am.LastActivity()
	if lastActivity != nil && lastActivity.IsBreak() == true {
		log.Println("Break is already set")
		return
	}

	a.am.SetBreak()
}

func (a App) HandleStatus() {

	defer a.OnError()

	now := time.Now()
	activities := a.am.Activities(now)
	if len(activities) == 0 {
		activities = a.am.Activities(now.AddDate(0, 0, -1))
	}

	if len(activities) == 0 {
		fmt.Println("Empty data")
		return
	}

	fmt.Println("ID\tTYPE\t\tSTART\t\tSTOP\t\tDATE")
	for k, v := range activities {
		fmt.Printf("%d\t%s\t\t%s\t%s\t%s\n",
			k, v.Type, v.StartString(), v.StopString(), v.DateString())
	}

}
