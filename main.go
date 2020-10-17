package main

import (
	"log"
	"math/rand"
	"os"
	"time"
)

const (
	ircHostURL  = "irc.twitch.tv"
	ircHostPort = "6667"
)

var (
	botNick      = os.Getenv("GDB_IRC_NICK")
	botPass      = os.Getenv("GDB_IRC_PASS")
	clientID     = os.Getenv("GDB_CLIENT_ID")
	clientSecret = os.Getenv("GDB_SECRET")
	dbInfo       = os.Getenv("GDB_DB_INFO")
	awsKeyID     = os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecret    = os.Getenv("AWS_SECRET_ACCESS_KEY")
)

func init() {
	rand.Seed(time.Now().Unix())
}

func run(irc *IrcConnection, dispatch *Dispatcher) {
	for {
		msg, err := irc.Read()
		if err != nil {
			log.Fatalln(err)
			if !irc.IsConnected() {
				log.Println("Attempting to reconnect...")
				irc.Connect(botNick, botPass)
			}
			continue
		}
		dispatch.Dispatch(msg)
	}
}

func main() {
	api, err := NewAPIClient(clientID, clientSecret)
	if err != nil {
		log.Fatalln(err)
	}

	db := NewDynamoConnection(awsKeyID, awsSecret)
	if err := db.Open(); err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	irc := NewIRCConnection(ircHostURL, ircHostPort)
	if err := irc.Connect(botNick, botPass); err != nil {
		log.Fatalln(err)
	}
	defer irc.Disconnect()

	// Look up which channels we're supposed to connect to
	connList, _ := db.GetRegisteredChannels()

	manager := NewChannelManager(connList, irc, api)
	dispatch := NewDispatcher(db, manager, api)

	run(irc, dispatch)
}
