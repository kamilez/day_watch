package main

import (
	"flag"
	"log"
	"os/user"

	"github.com/n4lik/day_watch/app"
	db "github.com/n4lik/day_watch/database"
	"github.com/n4lik/day_watch/utils"
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

var DatabasePath string

func init() {
	log.SetFlags(log.Llongfile)

	loginCmd = flag.Bool("login", false,
		"register login to system")
	logoutCmd = flag.Bool("logout", false,
		"register logout from system")
	graphCmd = flag.String("graph", "",
		"create graph image with working hours and breaks, enter the image path")
	breakCmd = flag.Bool("break", false,
		`set break from the next logout to the following login`)
	notification = flag.Bool("notify", false,
		`show time notification`)
	statusCmd = flag.Bool("status", false,
		`show working hours status`)

	flag.Parse()

	usr, err := user.Current()
	if err != nil {
		log.Fatalln(err.Error())
	}

	DatabasePath = usr.HomeDir + "/Documents/working_hours.db"
}

func main() {

	db := db.NewDatabase(DatabasePath)
	if db == nil {
		log.Fatalln("Could not create database")
	}

	notifier := utils.NewGnomeNotification("", "DayWatch")

	app := app.NewApp(db, notifier)

	if *notification == true {
		app.HandleNotification()
	} else if *loginCmd == true {
		app.HandleLogin()
	} else if *logoutCmd == true {
		app.HandleLogout()
	} else if *breakCmd == true {
		app.HandleBreak()
	} else if *statusCmd == true {
		app.HandleStatus()
	} else {
		app.HandleTomato()
	}

	errorMessage := recover()
	if errorMessage != nil {
		log.Fatalln(errorMessage)
	}
}
