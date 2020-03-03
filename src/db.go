package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"time"
)

const (
	logInsert     = "insert into logs (time, channel, username, message) values ($1, $2, $3, $4)"
	insertChannel = "insert into channels (name, id) values ($1, $2)"
	getChannels   = "select * from channels"
	deleteChannel = "DELETE FROM channels WHERE channels.name = $1"
)

func insertDB(db *sql.DB, time time.Time, channel string, username string, message string) {
	stmt, err := db.Prepare(logInsert)
	if err != nil {
		fmt.Println(err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(time, channel, username, message)
	if err != nil || res == nil {
		fmt.Println(err)
	}
}

func registerNewChannel(db *sql.DB, username string, id string) {
	stmt, err := db.Prepare(insertChannel)
	if err != nil {
		fmt.Println(err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(username, id)
	if err != nil || res == nil {
		fmt.Println(err)
	}
}

func removeChannel(db *sql.DB, username string) {
	stmt, err := db.Prepare(deleteChannel)
	if err != nil {
		fmt.Println(err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(username)
	if err != nil || res == nil {
		fmt.Println(err)
	}
}

func getRegisteredChannels(db *sql.DB) ([]string, []string) {
	rows, err := db.Query(getChannels)
	if err != nil {
		return nil, nil
	}

	defer rows.Close()

	channels := []string{}
	ids := []string{}

	for rows.Next() {
		var name, id string
		if err := rows.Scan(&name, &id); err != nil {
			return nil, nil
		}
		channels = append(channels, name)
		ids = append(ids, id)
		fmt.Println(name, id)
	}

	return channels, ids
}
