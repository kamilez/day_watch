package database

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/kamilez/day_watch/data"
	"github.com/kamilez/day_watch/utils"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db   *sql.DB
	path string
}

func NewDatabase(path string) *Database {

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Panic(err.Error())
	}

	return &Database{db, path}
}

func (db *Database) TableCreate(name string, obj interface{}) error {

	var query bytes.Buffer

	query.WriteString("CREATE TABLE IF NOT EXISTS '" + name + "' (id INTEGER PRIMARY KEY AUTOINCREMENT")

	typeOf := reflect.TypeOf(obj)
	for i := 0; i < typeOf.NumField(); i++ {
		query.WriteString(", " + typeOf.Field(i).Name + " " + typeOf.Field(i).Tag.Get("sql"))
	}
	query.WriteString(")")

	stmt, err := db.db.Prepare(query.String())
	if err != nil {
		log.Panic(err.Error())
	}

	_, err = stmt.Exec()
	if err != nil {
		log.Panic(err.Error())
	}

	return err
}

func (db *Database) RowAppend(name string, obj interface{}) error {

	var query bytes.Buffer

	query.WriteString("INSERT INTO '" + name + "' (")

	typeOf := reflect.TypeOf(obj)
	for i := 0; i < typeOf.NumField(); i++ {

		query.WriteString(typeOf.Field(i).Name)
		if i != typeOf.NumField()-1 {
			query.WriteString(", ")
		}
	}
	query.WriteString(") VALUES (")

	for i := 0; i < typeOf.NumField(); i++ {
		str := fmt.Sprintf("%v", reflect.ValueOf(obj).Field(i).Interface())

		switch reflect.ValueOf(obj).Field(i).Interface().(type) {
		case string:
			query.WriteString(`"` + str + `"`)
		default:
			query.WriteString(str)
		}

		if i != typeOf.NumField()-1 {
			query.WriteString(", ")
		}
	}
	query.WriteString(")")

	stmt, err := db.db.Prepare(query.String())
	if err != nil {
		log.Panic(err.Error())
	}

	_, err = stmt.Exec()
	if err != nil {
		log.Panic(err.Error())
	}

	return err
}

func (db *Database) FirstActivity(date string, activityType string) (data.Activity, error) {

	rows, err := db.db.Query(`
		SELECT start, stop, type
		FROM 'activities'
		WHERE date = ? AND type = ?
		ORDER BY ID LIMIT 1`,
		date, activityType,
	)
	if err != nil {
		log.Panic(err.Error())
	}
	defer rows.Close()

	activity := data.Activity{}
	for rows.Next() {
		err = rows.Scan(&activity.Start, &activity.Stop, &activity.Type)
		if err != nil {
			log.Panic(err.Error())
		}
	}

	return activity, nil
}

func (db *Database) LastActivity(date string, activityType string) (data.Activity, error) {

	rows, err := db.db.Query(
		`SELECT Start, Stop
		FROM 'activities' WHERE id = (SELECT MAX(ID)
		FROM 'activities' WHERE type = ? AND date = ?)`,
		activityType, date,
	)
	if err != nil {
		log.Panic(err.Error())
	}
	defer rows.Close()

	activity := data.Activity{Date: date, Type: activityType}
	for rows.Next() {
		err = rows.Scan(&activity.Start, &activity.Stop)
		if err != nil {
			log.Panic(err.Error())
		}

	}

	return activity, err
}

func (db *Database) UpdateActivityStartTime(activity data.Activity) error {

	query := fmt.Sprintf(
		`UPDATE 'activities'
		SET start = '%s'
		WHERE id = (SELECT MAX(ID)
		FROM 'activities' WHERE type = '%s' AND date = '%s')`,
		activity.Start, activity.Type, activity.Date,
	)

	return db.updateActivity(query, activity)
}

func (db *Database) UpdateActivityStopTime(activity data.Activity) error {

	query := fmt.Sprintf("UPDATE 'activities' SET stop = '%s' WHERE start = '%s' AND date = '%s'",
		activity.Stop, activity.Start, activity.Date)

	return db.updateActivity(query, activity)
}

func (db *Database) updateActivity(query string, activity data.Activity) error {

	stmt, err := db.db.Prepare(query)
	if err != nil {
		log.Panic(err.Error())
	}

	_, err = stmt.Exec()
	if err != nil {
		log.Panic(err.Error())
	}

	return nil
}

func (db *Database) hours(query string, date string) (time.Duration, error) {

	rows, err := db.db.Query(query, date)
	if err != nil {
		log.Panic(err.Error())
	}
	defer rows.Close()

	var duration time.Duration
	var start, stop string
	var startTime, stopTime time.Time

	for rows.Next() {
		err = rows.Scan(&start, &stop)
		if err != nil {
			log.Panic(err.Error())
		}

		if start == "" || stop == "" {
			continue
		}

		startTime, err = time.Parse(utils.DEFAULT_TIME_FORMAT, start)
		if err != nil {
			continue
		}

		stopTime, err = time.Parse(utils.DEFAULT_TIME_FORMAT, stop)
		if err != nil {
			continue
		}

		duration += stopTime.Sub(startTime)
	}

	return duration, nil
}

func (db *Database) SessionHours(date string) (time.Duration, error) {
	return db.hours("SELECT start, stop FROM 'activities' WHERE DATE = ? AND type = 'session'", date)
}

func (db *Database) BreakHours(date string) (time.Duration, error) {

	return db.hours("SELECT start, stop FROM 'activities' WHERE DATE = ? AND type = 'break'", date)
}

func (db *Database) Activities(date string) ([]data.Activity, error) {

	rows, err := db.db.Query(
		"SELECT type, start, stop FROM 'activities' WHERE date = ? ORDER BY id",
		date,
	)
	if err != nil {
		log.Panic(err.Error())
	}
	defer rows.Close()

	activities := make([]data.Activity, 0)

	for rows.Next() {

		activity := &data.Activity{}

		rows.Scan(&activity.Type, &activity.Start, &activity.Stop)
		activity.Date = date

		activities = append(activities, *activity)
	}

	return activities, nil
}
