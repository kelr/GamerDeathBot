package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/kelr/gundyr/helix"
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
	apiClient    *helix.Client
)

func init() {
	rand.Seed(time.Now().Unix())
}

func initHelixAPI() *helix.Client {
	config := &helix.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
	api, err := helix.NewClient(config)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return api
}

func main() {
	apiClient = initHelixAPI()

	db := NewDBConnection("postgres", dbInfo)
	if err := db.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer db.Close()

	irc := NewIRCConnection(ircHostURL, ircHostPort)
	if err := irc.Connect(botNick, botPass); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer irc.Disconnect()

	// Look up which channels we're supposed to connect to
	connList, _ := db.GetRegisteredChannels()
	manager := NewChannelManager(connList, irc)
	dispatcher := NewDispatcher(db, manager)

	for {
		msg, err := irc.Read()
		if err != nil {
			fmt.Println(err)
			if !irc.IsConnected() {
				fmt.Println("Attempting to reconnect...")
				time.Sleep(5 * time.Second)
				irc.Connect(botNick, botPass)
			}
			continue
		}
		dispatcher.Dispatch(msg)
	}
}
