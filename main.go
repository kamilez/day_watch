package main

import (
	"flag"
	"fmt"
	"log"
	"time"
)

var (
	loginCmd  *bool
	logoutCmd *bool

	imageCmd *string
	breakCmd *string
	graphCmd *string
)

type LogType string

const (
	NONE   LogType = "_"
	LOGIN  LogType = "I"
	LOGOUT LogType = "O"
)

type Notification struct {
	Type string `sql:"TEXT NOT NULL"`
	Time string `sql:"TEXT NOT NULL"`
	Date string `sql:"TEXT NOT NULL"`
}

type Session struct {
	Start  string `sql:"TEXT NOT NULL"`
	Stop   string `sql:"TEXT NOT NULL"`
	Date   string `sql:"TEXT NOT NULL"`
	Length int    `sql:"INT"`
}

const DATABASE_PATH = "/home/kamil/Documents/working_hours.db"
const DEFAULT_TIME_FORMAT = "15:04:05"
const DEFAULT_DATE_FORMAT = "02/01/2006"

var Db *Database

func init() {

	loginCmd = flag.Bool("login", false,
		"register login to system")
	logoutCmd = flag.Bool("logout", false,
		"register logout from system")
	graphCmd = flag.String("graph", "",
		"create graph image with working hours and breaks, enter the image path")
	breakCmd = flag.String("break", "",
		`set break hours in "hh:mm format"`)

	flag.Parse()
}

func main() {

	Db = NewDatabase(DATABASE_PATH)
	if Db == nil {
		log.Fatalln("Could not create database")
	}

	now := time.Now()
	noti := Notification{
		Type: "",
		Time: now.Format(DEFAULT_TIME_FORMAT),
		Date: now.Format(DEFAULT_DATE_FORMAT),
	}

	if *loginCmd == true {
		noti.Type = string(LOGIN)

		session := Session{
			Start:  now.Format(DEFAULT_TIME_FORMAT),
			Stop:   "",
			Date:   now.Format(DEFAULT_DATE_FORMAT),
			Length: 0,
		}

		Db.TableCreate("sessions", session)
		Db.RowAppend("sessions", session)

		total, err := Db.GetWorkedHours(session)
		if err != nil {
			log.Fatalln("Getting worked hours failed: ", err.Error())
		}

		firstActivity, err := Db.GetFirstActivity(session)
		if err != nil {
			log.Fatalln("Getting first activity failed: ", err.Error())
		}

		leaveTime := firstActivity.Add(8 * time.Hour)

		PostNotification(
			"Started:\t"+firstActivity.Format("15:04"),
			"Worked:\t"+total.Format("15:04"),
			"Leave:\t\t"+leaveTime.Format("15:04"),
		)

	} else if *logoutCmd == true {
		noti.Type = string(LOGOUT)

		session, err := Db.GetLastSession()
		if err != nil {
			log.Fatalln("Getting last session failed: ", err.Error())
		}

		if session.Stop == "" {

			start, err := time.Parse(DEFAULT_TIME_FORMAT, session.Start)
			if err != nil {
				log.Fatalln("Parsing string to time format failed: ", err.Error())
			}

			session.Stop = now.Format(DEFAULT_TIME_FORMAT)
			session.Length =
				now.Hour()*60 + now.Minute() - start.Hour()*60 - start.Minute()
			//session.Length = int(now.Sub(start))

			Db.UpdateSession(&session)
		}
	}

	Db.TableCreate("hours", noti)
	Db.RowAppend("hours", noti)

	if *breakCmd != "" {
	}
	if *graphCmd != "" {
	}
}

func PostNotification(info ...string) {

	var label string

	for _, v := range info {
		label += v + "\n"
	}

	notification := NewGnomeNotification("", "DayWatch", label)
	err := notification.Notify()
	if err != nil {
		fmt.Errorf("Posting notification failed:", err.Error())
	}
}
