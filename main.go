package main

import (
	"flag"
	"fmt"
	"log"
	"os/user"
	"time"
)

var (
	loginCmd     *bool
	logoutCmd    *bool
	breakCmd     *bool
	notification *bool
	statusCmd    *bool

	imageCmd *string
	graphCmd *string
)

type LogType string

const (
	NONE   LogType = "_"
	LOGIN  LogType = "I"
	LOGOUT LogType = "O"
)

type Activity struct {
	Start string `sql:"TEXT NOT NULL"`
	Stop  string `sql:"TEXT NOT NULL"`
	Date  string `sql:"TEXT NOT NULL"`
	Type  string `sql:"TEXT NOT NULL"`
}

const (
	DEFAULT_TIME_FORMAT = "15:04:05"
	DEFAULT_DATE_FORMAT = "02/01/2006"

	WORKING_HOURS_LENGTH = 8

	ACTIVITY_TYPE_SESSION = "session"
	ACTIVITY_TYPE_BREAK   = "break"
)

var IMAGE_PATH string
var DATABASE_PATH string
var Db *Database

func init() {

	loginCmd = flag.Bool("login", false,
		"register login to system")
	logoutCmd = flag.Bool("logout", false,
		"register logout from system")
	graphCmd = flag.String("graph", "",
		"create graph image with working hours and breaks, enter the image path")
	breakCmd = flag.Bool(ACTIVITY_TYPE_BREAK, false,
		`set break from the next logout to the following login`)
	notification = flag.Bool("notify", false,
		`show time notification`)
	statusCmd = flag.Bool("status", false,
		`show working hours status`)

	flag.Parse()

	usr, err := user.Current()
	ErrorCheck(err)

	DATABASE_PATH = usr.HomeDir + "/Documents/working_hours.db"
	IMAGE_PATH = usr.HomeDir + "/Documents/busy_beaver.png"
}

func main() {

	Db = NewDatabase(DATABASE_PATH)
	if Db == nil {
		log.Fatalln("Could not create database")
	}

	now := time.Now()
	nowTimeString := now.Format(DEFAULT_TIME_FORMAT)
	nowDateString := now.Format(DEFAULT_DATE_FORMAT)

	activity := Activity{
		Start: nowTimeString,
		Stop:  "",
		Date:  nowDateString,
	}

	Db.TableCreate("activities", activity)

	if *loginCmd == true || *notification == true {

		if *loginCmd == true {
			activity.Type = ACTIVITY_TYPE_SESSION
			lastBreak, err := Db.LastActivity(nowDateString, ACTIVITY_TYPE_BREAK)
			ErrorCheck(err)
			if lastBreak.Stop == "" {
				lastBreak.Stop = now.Format(DEFAULT_TIME_FORMAT)
				Db.UpdateActivityStopTime(lastBreak)
			}

			Db.RowAppend("activities", activity)
		}

		PostNotification(now)

	} else if *logoutCmd == true {

		lastActivity, err := Db.LastActivity(nowDateString, ACTIVITY_TYPE_SESSION)
		ErrorCheck(err)
		if lastActivity.Stop == "" {
			lastActivity.Stop = nowTimeString
			Db.UpdateActivityStopTime(lastActivity)
		}

		lastActivity, err = Db.LastActivity(nowDateString, ACTIVITY_TYPE_BREAK)
		ErrorCheck(err)
		if lastActivity.Start == "" {
			lastActivity.Start = nowTimeString
			Db.UpdateActivityStartTime(lastActivity)
		}

	} else if *breakCmd == true {
		activity.Type = ACTIVITY_TYPE_BREAK
		activity.Start = ""

		Db.RowAppend("activities", activity)
	} else if *statusCmd == true {
		activities, err := Db.Activities(nowDateString)
		ErrorCheck(err)

		if len(activities) == 0 {
			activities, err =
				Db.Activities(now.AddDate(0, 0, -1).Format(DEFAULT_DATE_FORMAT))
			ErrorCheck(err)
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
}

func PostNotification(now time.Time) {

	nowDate := now.Format(DEFAULT_DATE_FORMAT)

	breaks, err := Db.BreakHours(nowDate)
	ErrorCheck(err)

	firstActivity, err := Db.FirstActivity(nowDate, ACTIVITY_TYPE_SESSION)
	ErrorCheck(err)
	firstActivityStartTime, err := time.Parse(DEFAULT_TIME_FORMAT, firstActivity.Start)
	ErrorCheck(err)

	leaveTime := firstActivityStartTime.Add(WORKING_HOURS_LENGTH*time.Hour + breaks)
	total, err := Db.SessionHours(nowDate)
	ErrorCheck(err)

	if *notification == true {

		lastActivity, err := Db.LastActivity(nowDate, ACTIVITY_TYPE_SESSION)
		ErrorCheck(err)
		lastActivityTime, err := time.Parse(DEFAULT_TIME_FORMAT, lastActivity.Start)
		ErrorCheck(err)

		total += time.Duration(now.Hour()-lastActivityTime.Hour())*time.Hour +
			time.Duration(now.Minute()-lastActivityTime.Minute())*time.Minute
	}

	label := []string{
		"Started:\t" + firstActivityStartTime.Format("15:04"),
		"Leave:\t\t" + leaveTime.Format("15:04"),
		"Work time:\t" + fmt.Sprintf("%02d:%02d",
			int(total.Hours()), int(total.Minutes())%60),
	}

	notification := NewGnomeNotification("", "DayWatch", label...)
	notification.Notify()
}
