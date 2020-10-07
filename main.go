package main

import (
	"github.com/kelr/gundyr/helix"
	_ "github.com/lib/pq"
	"log"
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
	ircHostURL    = "irc.twitch.tv"
	ircHostPort   = "6667"
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
	apiClient  *helix.Client
)

// Parses out channel, username, and message strings from chat message
func splitMessage(msg string) (string, string, string) {
	return reChannel.FindString(msg), reUser.FindString(msg), reMessage.ReplaceAllLiteralString(msg, "")
}

func joinChannel(db *DBConnection, irc *IrcConnection, manager *ChannelManager, username string) {
	log.Println("JOIN -> " + username)
	botChannel, err := manager.GetChannel(botNick)
	if err != nil {
		log.Println(err)
		return
	}

	// Check if they have already registered
	if manager.IsRegistered(username) {
		botChannel.SendRegisterError(username)
		return
	}

	// Add a new DB entry, join the IRC channel, add the channel to the status map
	id := getChannelID(apiClient, username)
	if id == "" {
		log.Println("ERROR: API Can't get ID for: " + username)
		return
	}

	go db.AddChannel(username, id)
	irc.Join(username)
	manager.RegisterChannel(username, id, irc)

	botChannel.SendRegistered(username)
}

func leaveChannel(db *DBConnection, irc *IrcConnection, manager *ChannelManager, username string) {
	log.Println("LEAVE -> " + username)
	botChannel, err := manager.GetChannel(botNick)
	if err != nil {
		log.Println(err)
		return
	}

	// Check if they have already unregistered
	if !manager.IsRegistered(username) {
		botChannel.SendUnRegisterError(username)
		return
	}

	// Remove the DB entry, leave the IRC channel, delete the channel from the status map
	go db.DeleteChannelUser(username)
	irc.Part(username)
	manager.UnregisterChannel(username)

	botChannel.SendUnregistered(username)
}

// Determines the command from the chat message, if any, and executes it
func parseMessage(db *DBConnection, irc *IrcConnection, manager *ChannelManager, channel string, username string, message string) {

	if channel == "#"+botNick {
		// Handle home channel commands
		if message == "!join" {
			joinChannel(db, irc, manager, username)
		} else if message == "!leave" {
			leaveChannel(db, irc, manager, username)
		}
	} else {
		chatChan, err := manager.GetChannel(channel[1:])
		if err != nil {
			log.Println(err)
			return
		}
		if reGreeting.FindString(message) != "" {
			chatChan.SendGreeting(username)
		} else if reFarewell.FindString(message) != "" {
			chatChan.SendFarewell(username)
		} else if message == "!gamerdeath" {
			chatChan.SendGamerdeath()
		}
	}
}

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
		log.Println(err)
		os.Exit(1)
	}
	return api
}

func main() {
	apiClient = initHelixAPI()

	db := NewDBConnection("postgres", dbInfo)
	if err := db.Open(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer db.Close()

	irc := NewIRCConnection(ircHostURL, ircHostPort)
	if err := irc.Connect(botNick, botPass); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer irc.Disconnect()

	// Look up which channels we're supposed to connect to
	connList, idList := db.GetRegisteredChannels()
	for _, channel := range connList {
		irc.Join(channel)
	}

	manager := NewChannelManager(connList, idList, irc)

	// Main thread rxs on connection, logs to db and responds
	for {
		msg, err := irc.Recv()
		if err != nil {
			log.Println(err)
			log.Println("Attempting to reconnect...")
			irc.Connect(botNick, botPass)
		}
		channel, username, message := splitMessage(msg)
		if username != "tmi" && username != botNick {
			// Log out the message to the db
			go db.InsertLog(time.Now(), channel, username, message)
			parseMessage(db, irc, manager, channel, username, message)
		}
	}

}
