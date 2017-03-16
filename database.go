package main

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"reflect"
	"time"
)

type Database struct {
	db   *sql.DB
	path string
}

func NewDatabase(path string) *Database {

	db, err := sql.Open("sqlite3", path)
	ErrorCheck(err)

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

	log.Println("Table create query: ", query.String())

	stmt, err := db.db.Prepare(query.String())
	ErrorCheck(err)

	_, err = stmt.Exec()
	ErrorCheck(err)

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

	log.Println("Query: ", query.String())

	stmt, err := db.db.Prepare(query.String())
	ErrorCheck(err)

	_, err = stmt.Exec()
	ErrorCheck(err)

	return err
}

func (db *Database) FirstActivity(date string, activityType string) (Activity, error) {

	rows, err := db.db.Query(`
		SELECT start, stop, type
		FROM 'activities'
		WHERE date = ? AND type = ?
		ORDER BY ID LIMIT 1`,
		date, activityType,
	)
	ErrorCheck(err)
	defer rows.Close()

	activity := Activity{}
	for rows.Next() {
		err = rows.Scan(&activity.Start, &activity.Stop, &activity.Type)
		ErrorCheck(err)
	}

	return activity, nil
}

func (db *Database) LastActivity(date string, activityType string) (Activity, error) {

	rows, err := db.db.Query(
		`SELECT Start, Stop
		FROM 'activities' WHERE id = (SELECT MAX(ID)
		FROM 'activities' WHERE type = ? AND date = ?)`,
		activityType, date,
	)
	ErrorCheck(err)
	defer rows.Close()

	activity := Activity{Date: date, Type: activityType}
	for rows.Next() {
		err = rows.Scan(&activity.Start, &activity.Stop)
		ErrorCheck(err)

	}

	return activity, err
}

func (db *Database) UpdateActivityStartTime(activity Activity) error {

	query := fmt.Sprintf(
		`UPDATE 'activities'
		SET start = '%s'
		WHERE id = (SELECT MAX(ID)
		FROM 'activities' WHERE type = '%s' AND date = '%s')`,
		activity.Start, activity.Type, activity.Date,
	)

	return Db.updateActivity(query, activity)
}

func (db *Database) UpdateActivityStopTime(activity Activity) error {

	query := fmt.Sprintf("UPDATE 'activities' SET stop = '%s' WHERE start = '%s' AND date = '%s'",
		activity.Stop, activity.Start, activity.Date)

	return Db.updateActivity(query, activity)
}

func (db *Database) updateActivity(query string, activity Activity) error {

	log.Println("Update session query: ", query)

	stmt, err := db.db.Prepare(query)
	ErrorCheck(err)

	_, err = stmt.Exec()
	ErrorCheck(err)

	return nil
}

func (db *Database) hours(query string, date string) (time.Duration, error) {

	rows, err := db.db.Query(query, date)
	ErrorCheck(err)
	defer rows.Close()

	var duration time.Duration
	var start, stop string
	var startTime, stopTime time.Time

	for rows.Next() {
		err = rows.Scan(&start, &stop)
		ErrorCheck(err)

		if start == "" || stop == "" {
			continue
		}

		startTime, err = time.Parse(DEFAULT_TIME_FORMAT, start)
		if err != nil {
			continue
		}

		stopTime, err = time.Parse(DEFAULT_TIME_FORMAT, stop)
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

func (db *Database) Activities(date, typeOf string) ([]Activity, error) {

	rows, err := Db.db.Query(
		"SELECT start, stop FROM 'activities' WHERE type = ? AND date = ? ORDER BY id",
		typeOf, date,
	)
	ErrorCheck(err)
	defer rows.Close()

	activities := make([]Activity, 0)

	for rows.Next() {

		activity := &Activity{}

		activity.Type = typeOf
		activity.Date = date
		rows.Scan(&activity.Start, &activity.Stop)

		activities = append(activities, *activity)
	}

	log.Println(activities)

	return activities, nil
}
