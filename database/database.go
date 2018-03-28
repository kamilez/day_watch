package database

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"

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

	database := &Database{db, path}
	database.createActivityTable()

	return database
}

func (db *Database) createActivityTable() {

	query := `CREATE TABLE IF NOT EXISTS 'activities'
		(id INTEGER PRIMARY KEY AUTOINCREMENT,
		start TEXT,
		stop TEXT,
		type TEXT NOT NULL,
		weekday TEXT)`

	stmt, err := db.db.Prepare(query)
	if err != nil {
		log.Panic(err.Error())
	}

	_, err = stmt.Exec()
	if err != nil {
		log.Panic(err.Error())
	}
}

func (db *Database) AppendActivityRow(activity *data.Activity) error {

	var query bytes.Buffer

	query.WriteString("INSERT INTO 'activities' (start, stop, type, weekday) VALUES (")
	query.WriteString("'" + utils.FormattedDatetime(activity.Start) + "', ")
	query.WriteString("'" + utils.FormattedDatetime(activity.Stop) + "', ")
	query.WriteString("'" + string(activity.Type) + "',")
	query.WriteString("'" + activity.Weekday() + "')")

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

func (db *Database) UpdateActivityStartTime(activity data.Activity) {

	query := fmt.Sprintf(`UPDATE 'activities'
		SET start = '%s'
		WHERE id = (SELECT MAX(ID)
		FROM 'activities' WHERE type = '%s')`,
		utils.FormattedDatetime(activity.Start), activity.Type,
	)

	db.updateActivity(query, activity)
}

func (db *Database) UpdateActivityStopTime(activity data.Activity) {

	query := fmt.Sprintf("UPDATE 'activities' SET stop = '%s' WHERE start = '%s'",
		utils.FormattedDatetime(activity.Stop), utils.FormattedDatetime(activity.Start))

	db.updateActivity(query, activity)
}

func (db *Database) updateActivity(query string, activity data.Activity) {

	stmt, err := db.db.Prepare(query)
	if err != nil {
		log.Panic(err.Error())
	}

	_, err = stmt.Exec()
	if err != nil {
		log.Panic(err.Error())
	}
}

func (db *Database) activities(query string, args ...interface{}) []data.Activity {

	rows, err := db.db.Query(query, args...)
	if err != nil {
		log.Panic(err.Error())
	}
	defer rows.Close()

	activities := make([]data.Activity, 0)

	var start, stop, aType string
	for rows.Next() {

		err = rows.Scan(&start, &stop, &aType)
		if err != nil {
			log.Panic(err.Error())
		}

		activity := &data.Activity{
			Start: *utils.String2Time(start),
			Stop:  *utils.String2Time(stop),
			Type:  data.ActivityType(aType),
		}

		activities = append(activities, *activity)
	}

	return activities
}

func (db *Database) Activities(since, typeOf string) []data.Activity {
	if typeOf == "session" || typeOf == "break" {
		return db.activities(`SELECT start, stop, type FROM 'activities'
			WHERE type = ? and start >= date(?, 'start of day')
			ORDER BY ID`,
			typeOf, since)
	}

	return db.activities(`SELECT start, stop, type FROM 'activities'
		WHERE start >= date(?, 'start of day')
		ORDER BY ID`,
		since)

}

func (db *Database) FirstActivity(since, typeOf string) *data.Activity {

	var result data.Activity

	activities := db.Activities(since, typeOf)

	if len(activities) > 0 {
		result = activities[0]
		return &result
	}

	return nil
}

func (db *Database) LastActivity(typeOf string) *data.Activity {

	var activities []data.Activity

	if typeOf == "session" || typeOf == "break" {
		activities = db.activities(`SELECT start, stop, type FROM 'activities'
			WHERE id = (SELECT MAX(id) FROM 'activities'
			WHERE type = ?)`,
			typeOf)
	} else {
		activities = db.activities(`SELECT start, stop, type FROM 'activities'
			WHERE id = (SELECT MAX(id) FROM 'activities')`)
	}

	if len(activities) != 0 {
		activity := activities[0]
		return &activity
	}

	return nil
}
