package main

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"reflect"
	"time"
)

type Database struct {
	db   *sql.DB
	path string
}

func NewDatabase(path string) *Database {

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		fmt.Errorf("Creating database failed:", err.Error())
		return nil
	}

	return &Database{db, path}
}

func (db *Database) TableCreate(name string, obj interface{}) error {

	_type := reflect.TypeOf(obj)

	var query bytes.Buffer
	query.WriteString("CREATE TABLE IF NOT EXISTS '" + name + "' (id INTEGER PRIMARY KEY AUTOINCREMENT")

	for i := 0; i < _type.NumField(); i++ {

		query.WriteString(", " + _type.Field(i).Name + " " + _type.Field(i).Tag.Get("sql"))
	}

	query.WriteString(")")

	fmt.Println("Query: ", query.String())

	stmt, err := db.db.Prepare(query.String())
	if err != nil {
		fmt.Errorf("Table creation failed:", err.Error())
		return err
	}

	_, err = stmt.Exec()

	return err
}

func (db *Database) RowAppend(name string, obj interface{}) error {

	_type := reflect.TypeOf(obj)

	var query bytes.Buffer

	query.WriteString("INSERT INTO '" + name + "' (")

	for i := 0; i < _type.NumField(); i++ {

		query.WriteString(_type.Field(i).Name)
		if i != _type.NumField()-1 {
			query.WriteString(", ")
		}
	}

	query.WriteString(") VALUES (")

	for i := 0; i < _type.NumField(); i++ {

		str := fmt.Sprintf("%v", reflect.ValueOf(obj).Field(i).Interface())

		switch reflect.ValueOf(obj).Field(i).Interface().(type) {
		case string:
			query.WriteString(`"` + str + `"`)
		default:
			query.WriteString(str)
		}

		if i != _type.NumField()-1 {
			query.WriteString(", ")
		}
	}

	query.WriteString(")")

	stmt, err := db.db.Prepare(query.String())
	fmt.Println("Query: ", query.String())
	if err != nil {
		fmt.Errorf("Table creation failed:", err.Error())
		return err
	}

	_, err = stmt.Exec()

	return err
}

func (db *Database) GetLastNotification() (Notification, error) {

	query := "SELECT * FROM 'hours' WHERE ID = (SELECT MAX(ID) FROM 'hours')"
	fmt.Println("Query: ", query)

	rows, err := db.db.Query(query)
	defer rows.Close()
	if err != nil {
		fmt.Errorf("Error: ", err.Error())
		return Notification{}, err
	}

	noti := Notification{}
	for rows.Next() {
		var id int
		err = rows.Scan(&id, &noti.Type, &noti.Time)
		if err != nil {
			fmt.Errorf("Error: ", err.Error())
			return Notification{}, err
		}
	}

	return noti, err
}

func (db *Database) GetLastSession() (Session, error) {

	query := "SELECT * FROM 'sessions' WHERE ID = (SELECT MAX(ID) FROM 'sessions')"
	fmt.Println("Query: ", query)

	rows, err := db.db.Query(query)
	defer rows.Close()
	if err != nil {
		fmt.Errorf("Error: ", err.Error())
		return Session{}, err
	}

	session := Session{}
	for rows.Next() {
		var id int
		err = rows.Scan(&id, &session.Start, &session.Stop, &session.Date, &session.Length)
		if err != nil {
			fmt.Errorf("Error: %s", err.Error())
			return Session{}, err
		}
	}

	return session, err
}

func (db *Database) UpdateSession(session *Session) error {

	query := fmt.Sprintf("UPDATE 'sessions' SET STOP = '%s', LENGTH = '%d' WHERE START = '%s' AND DATE = '%s'",
		session.Stop, session.Length, session.Start, session.Date)

	fmt.Println(query)

	stmt, err := db.db.Prepare(query)
	if err != nil {
		fmt.Println("Query preparation failed")
		fmt.Errorf(err.Error())
		return err
	}

	_, err = stmt.Exec()
	if err != nil {
		fmt.Println("Query execution failed")
		fmt.Errorf("", err.Error())
		return err
	}

	return nil
}

func (db *Database) GetWorkedHours(session Session) (time.Time, error) {

	query := fmt.Sprintf("SELECT * FROM 'sessions' WHERE DATE = '%s'", session.Date)

	rows, err := db.db.Query(query)
	defer rows.Close()
	if err != nil {
		fmt.Errorf("Error: ", err.Error())
		return time.Time{}, err
	}

	var totalMinutes int

	sn := Session{}
	var id int

	for rows.Next() {
		err = rows.Scan(&id, &sn.Start, &sn.Stop, &sn.Date, &sn.Length)
		fmt.Println(sn.Length)
		if err != nil {
			fmt.Errorf("Error: %s", err.Error())
			return time.Time{}, err
		}

		totalMinutes += sn.Length
	}

	return time.Time{}.Add(time.Duration(totalMinutes) * time.Minute), nil
}

func (db *Database) GetFirstActivity(session Session) (time.Time, error) {

	query := fmt.Sprintf("SELECT * FROM 'sessions' WHERE DATE = '%s' ORDER BY ID LIMIT 1", session.Date)
	fmt.Println("Query: ", query)

	rows, err := db.db.Query(query)
	defer rows.Close()
	if err != nil {
		fmt.Errorf("Error: ", err.Error())
		return time.Time{}, err
	}

	sn := Session{}
	var id int

	for rows.Next() {
		err = rows.Scan(&id, &sn.Start, &sn.Stop, &sn.Date, &sn.Length)
		if err != nil {
			fmt.Println("asdsad")
			fmt.Errorf("Error: %s", err.Error())
			return time.Time{}, err
		}
	}

	formattedTime, err := time.Parse(DEFAULT_TIME_FORMAT, sn.Start)
	if err != nil {
		fmt.Println("asdsadasdsada")
		fmt.Errorf("Error: ", err.Error())
		return time.Time{}, err
	}

	return formattedTime, nil
}
