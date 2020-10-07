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
	deleteChannelUser = "DELETE FROM channels WHERE channels.name = $1"
	deleteChannelID = "DELETE FROM channels WHERE channels.id = $1"
)

// DB is an interface to the underlying database connection object
type DB interface {
	Close() error
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

// Database is a wrapper over a lower level database driver
type Database interface {
	Open() error
	Close() error
	InsertLog(time time.Time, channel string, username string, message string) error
	AddChannel(username string, id string)
	DeleteChannelUser(username string)
	DeleteChannelID(id string)
}

// DBConnection represents a connection state to the channel Database Server
type DBConnection struct {
	conn   DB
	driver string
	info string
}

// NewDBConnection returns a DBConnection object
func NewDBConnection(driver string, info string) *DBConnection {
	return &DBConnection{
		conn:   nil,
		driver: driver,
		info:   info,
	}
}

// Open initializes a new database connection
func (db *DBConnection) Open() error {
	dbConn, err := sql.Open(db.driver, db.info)
	if err != nil {
		return err
	}
	db.conn = dbConn
	return nil
}

// Close closes the database connection
func (db *DBConnection) Close() error {
	if err := db.conn.Close(); err != nil {
		return err
	}
	return nil
}

// AddChannel registers a new channel
func (db *DBConnection) AddChannel(username string, id string) {
	stmt, err := db.conn.Prepare(insertChannel)
	if err != nil {
		fmt.Println(err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(username, id)
	if err != nil || res == nil {
		fmt.Println(err)
	}
}

// DeleteChannelUser deletes a registered channel by username
func (db *DBConnection) DeleteChannelUser(username string) {
	stmt, err := db.conn.Prepare(deleteChannelUser)
	if err != nil {
		fmt.Println(err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(username)
	if err != nil || res == nil {
		fmt.Println(err)
	}
}

// DeleteChannelID deletes a registered channel by id
func (db *DBConnection) DeleteChannelID(id string) {
	stmt, err := db.conn.Prepare(deleteChannelID)
	if err != nil {
		fmt.Println(err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(id)
	if err != nil || res == nil {
		fmt.Println(err)
	}
}

// InsertLog inserts a log
func (db *DBConnection) InsertLog(time time.Time, channel string, username string, message string) {
	stmt, err := db.conn.Prepare(logInsert)
	if err != nil {
		fmt.Println(err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(time, channel, username, message)
	if err != nil || res == nil {
		fmt.Println(err)
	}
}

// GetRegisteredChannels returns two slices of channelnames and ids that are registered
func (db *DBConnection) GetRegisteredChannels() ([]string, []string) {
	rows, err := db.conn.Query(getChannels)
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
