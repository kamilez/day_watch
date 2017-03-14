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

func (db *Database) GetLastNotification() (Notification, error) {

	query := "SELECT * FROM 'hours' WHERE ID = (SELECT MAX(ID) FROM 'hours')"
	log.Println("Query: ", query)

	rows, err := db.db.Query(query)
	ErrorCheck(err)

	defer rows.Close()

	noti := Notification{}
	var id int
	for rows.Next() {
		err = rows.Scan(&id, &noti.Type, &noti.Time)
		ErrorCheck(err)
	}

	return noti, err
}

func (db *Database) GetLastSession() (Session, error) {

	query := "SELECT * FROM 'sessions' WHERE ID = (SELECT MAX(ID) FROM 'sessions')"
	log.Println("Select last session query: ", query)

	rows, err := db.db.Query(query)
	ErrorCheck(err)
	defer rows.Close()

	session := Session{}
	var id int
	for rows.Next() {
		err = rows.Scan(&id, &session.Start, &session.Stop, &session.Date, &session.Length)
		ErrorCheck(err)
	}

	return session, err
}

func (db *Database) UpdateSession(session *Session) error {

	query := fmt.Sprintf("UPDATE 'sessions' SET STOP = '%s', LENGTH = '%d' WHERE START = '%s' AND DATE = '%s'",
		session.Stop, session.Length, session.Start, session.Date)

	log.Println("Update session query: ", query)

	stmt, err := db.db.Prepare(query)
	ErrorCheck(err)

	_, err = stmt.Exec()
	ErrorCheck(err)

	return nil
}

func (db *Database) GetWorkedHours(session Session) (time.Time, error) {

	rows, err := db.db.Query("SELECT * FROM 'sessions' WHERE DATE = ?", session.Date)
	ErrorCheck(err)
	defer rows.Close()

	sn := Session{}
	var id, totalMinutes int

	for rows.Next() {
		err = rows.Scan(&id, &sn.Start, &sn.Stop, &sn.Date, &sn.Length)
		ErrorCheck(err)

		totalMinutes += sn.Length
	}

	return time.Time{}.Add(time.Duration(totalMinutes) * time.Minute), nil
}

func (db *Database) GetFirstActivity(session Session) (time.Time, error) {

	rows, err := db.db.Query("SELECT * FROM 'sessions' WHERE DATE = ? ORDER BY ID LIMIT 1",
		session.Date)
	ErrorCheck(err)
	defer rows.Close()

	sn := Session{}
	var id int

	for rows.Next() {
		err = rows.Scan(&id, &sn.Start, &sn.Stop, &sn.Date, &sn.Length)
		ErrorCheck(err)
	}

	formattedTime, err := time.Parse(DEFAULT_TIME_FORMAT, sn.Start)
	ErrorCheck(err)

	return formattedTime, nil
}
