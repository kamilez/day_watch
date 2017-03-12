package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

var (
	loginCmd  *bool
	logoutCmd *bool

	imageCmd *string
	breakCmd *string
	graphCmd *string
)

type Notification struct {
	Type   string `sql:"TEXT NOT NULL"`
	Hour   int    `sql:"INT"`
	Minute int    `sql:"INT"`
	Day    int    `sql:"INT"`
	Month  int    `sql:"INT"`
	Year   int    `sql:"INT"`
}

const DATABASE_PATH = "/home/kamil/Documents/working_hours.db"

var Db *Database

func init() {

	loginCmd = flag.Bool("login", false, "register login to system")
	logoutCmd = flag.Bool("logout", false, "register logout from system")

	graphCmd = flag.String("graph", "", "create graph image with working hours and breaks, enter the image path")
	breakCmd = flag.String("break", "", `set break hours in "hh:mm format"`)
	imageCmd = flag.String("image", "", "set notification image path")

	flag.Parse()

	Db = NewDatabase(DATABASE_PATH)
	if Db == nil {
		os.Exit(1)
	}
}

func main() {

	now := time.Now()

	noti := Notification{
		"",
		now.Hour(),
		now.Minute(),
		now.Day(),
		int(now.Month()),
		int(now.Year()),
	}

	if *loginCmd == true {
		noti.Type = "Login"
		handleLoginCmd(noti)

		current := timeToSimpleFormat(now)
		leave := timeToSimpleFormat(now.Add(time.Hour * 8))

		notification := NewGnomeNotification("", "DayWatch", "Login: "+current+"\nLeave: "+leave, "")
		err := notification.Notify()
		if err != nil {
			fmt.Errorf("Posting notification failed:", err.Error())
		}

	} else if *logoutCmd == true {
		noti.Type = "Logout"
		handleLogoutCmd(noti)
	}

	if *breakCmd != "" {
		//Parse and handle hour
	}
	if *graphCmd != "" {

	}
}

func handleLoginCmd(noti Notification) {

	Db.TableCreate("hours", noti)
	Db.RowAppend("hours", noti)
}

func handleLogoutCmd(noti Notification) {

	Db.TableCreate("hours", noti)
	Db.RowAppend("hours", noti)

	noti, _ = Db.RowGetLast("hours")
	fmt.Println("last", noti)
}

func timeToSimpleFormat(t time.Time) string {

	format := fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute())

	return format
}
