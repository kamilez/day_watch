package main

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"reflect"
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

func (db *Database) RowGetLast(name string) (Notification, error) {

	query := "SELECT * FROM '" + name + "' WHERE ID = (SELECT MAX(ID) FROM '" + name + "')"
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
		err = rows.Scan(&id, &noti.Type, &noti.Hour, &noti.Minute, &noti.Day, &noti.Month, &noti.Year)
		if err != nil {
			fmt.Errorf("Error: ", err.Error())
			return Notification{}, err
		}
	}

	return noti, err
}
