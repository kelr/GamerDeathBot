package main

import (
	"database/sql"
	"fmt"
	"github.com/kelr/gundyr/helix"
	_ "github.com/lib/pq"
	"math/rand"
	"os"
	"regexp"
	"time"
)

const (
	regexUsername = `\w+`
	regexChannel  = `#\w+`
	regexMessage  = `^:\w+!\w+@\w+\.tmi\.twitch\.tv PRIVMSG #\w+ :`
	regexGreeting = `(?i)(hi|hiya|hello|hey|yo|sup|howdy|hovvdy|greetings|what's good|whats good|vvhat's good|vvhats good|what's up|whats up|vvhat's up|vvhats up|konichiwa|hewwo|etalWave|vvhats crackalackin|whats crackalackin|henlo|good morning|good evening|good afternoon) (@*GamerDeathBot|gdb)`
	regexFarewell = `(?i)(bye|goodnight|good night|goodbye|good bye|see you|see ya|so long|farewell|later|seeya|ciao|au revoir|bon voyage|peace|in a while crocodile|see you later alligator|later alligator|have a good one|igottago|l8r|later skater|catch you on the flip side|bye-bye|sayonara) (@*GamerDeathBot|gdb)`
)

var (
	botNick      = os.Getenv("GDB_IRC_NICK")
	botPass      = os.Getenv("GDB_IRC_PASS")
	clientID     = os.Getenv("GDB_CLIENT_ID")
	clientSecret = os.Getenv("GDB_SECRET")
	dbInfo       = os.Getenv("GDB_DB_INFO")

	reUser     = regexp.MustCompile(regexUsername)
	reChannel  = regexp.MustCompile(regexChannel)
	reMessage  = regexp.MustCompile(regexMessage)
	reGreeting = regexp.MustCompile(regexGreeting)
	reFarewell = regexp.MustCompile(regexFarewell)
	apiClient *helix.Client
)

// Parses out channel, username, and message strings from chat message
func splitMessage(msg string) (string, string, string) {
	return reChannel.FindString(msg), reUser.FindString(msg), reMessage.ReplaceAllLiteralString(msg, "")
}

func joinChannel(db *sql.DB, irc *IrcConnection, channelTransmit *map[string]*ChatChannel, username string) {
	fmt.Println("JOIN -> " + username)
	botChannel := (*channelTransmit)[botNick]

	// Check if they have already registered
	if _, ok := (*channelTransmit)[username]; ok {
		botChannel.SendRegisterError(username)
	}

	// Add a new DB entry, join the IRC channel, add the channel to the status map
	id := getChannelID(apiClient, username)
	if id == "" {
		fmt.Println("ERROR: API Can't get ID for: " + username)
	}

	go registerNewDBChannel(db, username, id)
	irc.Join(username)
	(*channelTransmit)[username] = NewChatChannel(username, id, irc)
	go (*channelTransmit)[username].StartGetupTimer()

	botChannel.SendRegistered(username)
}

func leaveChannel(db *sql.DB, irc *IrcConnection, channelTransmit *map[string]*ChatChannel, username string) {
	fmt.Println("LEAVE -> " + username)
	botChannel := (*channelTransmit)[botNick]

	// Check if they have already unregistered
	if _, ok := (*channelTransmit)[username]; !ok {
		botChannel.SendUnRegisterError(username)
		return
	}

	// Remove the DB entry, leave the IRC channel, delete the channel from the status map
	go removeDBChannel(db, username)
	irc.Part(username)
	(*channelTransmit)[username].StopGetupTimer()
	delete((*channelTransmit), username)

	botChannel.SendUnregistered(username)
}

// Determines the command from the chat message, if any, and executes it
func parseMessage(db *sql.DB, irc *IrcConnection, channelTransmit *map[string]*ChatChannel, channel string, username string, message string) {
	if message == "!join" && channel == "#"+botNick {
		joinChannel(db, irc, channelTransmit, username)
	} else if message == "!leave" && channel == "#"+botNick {
		leaveChannel(db, irc, channelTransmit, username)
	} else if reGreeting.FindString(message) != "" {
		(*channelTransmit)[channel[1:]].SendGreeting(username)
	} else if reFarewell.FindString(message) != "" {
		(*channelTransmit)[channel[1:]].SendFarewell(username)
	} else if message == "!gamerdeath" {
		(*channelTransmit)[channel[1:]].SendGamerdeath()
	}
}

func init() {
	rand.Seed(time.Now().Unix())
}

func main() {
	if clientID != "y3s7zuy35l3nvrcxhd4lvozw5ycnb6" {
		fmt.Println("[" +  clientID + "]")
	}
	config := &helix.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
	tmp, err := helix.NewClient(config)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	apiClient = tmp

	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	connList, idList := getRegisteredChannels(db)

	irc := NewIRCConnection(connList)
	if irc.Connect(botNick, botPass) != nil {
		fmt.Println(err)
		return
	}
	defer irc.Disconnect()

	// Register channels into the transmission map and start their timers
	channelTransmit := make(map[string]*ChatChannel)
	for index, channel := range connList {
		channelTransmit[channel] = NewChatChannel(channel, idList[index], irc)
		go channelTransmit[channel].StartGetupTimer()
	}

	// Main thread rxs on connection, logs to db and responds
	for {
		msg, err := irc.Recv()
		if err != nil {
			fmt.Println("Attempting to reconnect...")
			irc.Connect(botNick, botPass)
		}
		channel, username, message := splitMessage(msg)
		if username != "tmi" && username != botNick {
			//fmt.Println(time.Now().Format(time.StampMilli), ":", channel, "-", username, "-", message)
			// Log out the message to the db
			go insertDB(db, time.Now(), channel, username, message)
			parseMessage(db, irc, &channelTransmit, channel, username, message)
		}
	}
}
