package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"regexp"
	"time"
)

const (
	botNick       = ""
	botPass       = ""
	clientID      = ""
    dbInfo        = ""
	regexUsername = `\w+`
	regexChannel  = `#\w+`
	regexMessage  = `^:\w+!\w+@\w+\.tmi\.twitch\.tv PRIVMSG #\w+ :`
	regexGreeting = `(?i)(hi|hiya|hello|hey|yo|sup|howdy|hovvdy|greetings|what's good|whats good|vvhat's good|vvhats good|what's up|whats up|vvhat's up|vvhats up|konichiwa|hewwo|etalWave|vvhats crackalackin|whats crackalackin|henlo|good morning|good evening|good afternoon) (@*GamerDeathBot|gdb)`
	regexFarewell = `(?i)(bye|goodnight|good night|goodbye|good bye|see you|see ya|so long|farewell|later|seeya|ciao|au revoir|bon voyage|peace|in a while crocodile|see you later alligator|later alligator|have a good one|igottago|l8r|later skater|catch you on the flip side|bye-bye|sayonara) (@*GamerDeathBot|gdb)`
)

var reUser = regexp.MustCompile(regexUsername)
var reChannel = regexp.MustCompile(regexChannel)
var reMessage = regexp.MustCompile(regexMessage)
var reGreeting = regexp.MustCompile(regexGreeting)
var reFarewell = regexp.MustCompile(regexFarewell)

// Parses out channel, username, and message strings from chat message
func splitMessage(msg string) (string, string, string) {
	return reChannel.FindString(msg), reUser.FindString(msg), reMessage.ReplaceAllLiteralString(msg, "")
}

func parseMessage(db *sql.DB, channelMap *map[string]*ChatChannel, channel string, username string, message string) {
	if message == "!join" && channel == "#gamertestbot" {
		fmt.Println("REGISTERED " + username)
		// TODO: api query for id
		registerNewChannel(db, username, "31903323")
	} else if message == "!leave" && channel == "#gamertestbot" {
		fmt.Println("UNREGISTERED " + username)
		removeChannel(db, username)
	} else if reGreeting.FindString(message) != "" {
		currChannel := (*channelMap)[channel[1:]]
		currChannel.SendGreeting(username)
	} else if reFarewell.FindString(message) != "" {
		currChannel := (*channelMap)[channel[1:]]
		currChannel.SendFarewell(username)
	} else if message == "!gamerdeath" {
		currChannel := (*channelMap)[channel[1:]]
		currChannel.SendGamerdeath()
	}
}

func main() {
	//api := twitchapi.NewTwitchClient(clientID)
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	connList := getRegisteredChannels(db)

	irc := NewIRCConnection(connList)
	irc.Connect(botNick, botPass)
	defer irc.Disconnect()

	channelMap := make(map[string]*ChatChannel)
	for _, channel := range connList {
		// TODO channel id
		channelMap[channel] = NewChatChannel(channel, "123", irc)
		fmt.Println("Registered: " + channel)
	}

	for {
		msg, _ := irc.Recv()
		channel, username, message := splitMessage(msg)
		if username != "tmi" && username != botNick {
			//fmt.Println(time.Now().Format(time.StampMilli), ":", channel, "-", username, "-", message)
			go insertDB(db, time.Now(), channel, username, message)
			parseMessage(db, &channelMap, channel, username, message)
		}
	}
}
