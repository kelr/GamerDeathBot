package main

import (
	"fmt"
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
	reUser     = regexp.MustCompile(regexUsername)
	reChannel  = regexp.MustCompile(regexChannel)
	reMessage  = regexp.MustCompile(regexMessage)
	reGreeting = regexp.MustCompile(regexGreeting)
	reFarewell = regexp.MustCompile(regexFarewell)
)

// Parser determines what to execute per chat message
type Parser struct {
	db      Database
	manager *ChannelManager
}

// IRCMessage represents parsed out fields of a message from IRC
type IRCMessage struct {
	Channel  string
	Username string
	Message  string
	Tags     IRCTags
}

// IRCTags represents parsed out IRC tags
type IRCTags struct {
	tmp string
}

// Parse parses out commands from a chat message
func (p *Parser) Parse(msg string) {
	ircMessage := IRCMessage{
		Channel:  reChannel.FindString(msg),
		Username: reUser.FindString(msg),
		Message:  reMessage.ReplaceAllLiteralString(msg, ""),
	}
	p.Dispatch(ircMessage)
}

// NewParser returns a new Parser
func NewParser(db Database, manager *ChannelManager) *Parser {
	return &Parser{
		db:      db,
		manager: manager,
	}
}

// Dispatch determines what command to run based on the input IRC Message
func (p *Parser) Dispatch(msg IRCMessage) {
	// Don't reply to self or TMI IRC messages
	if msg.Username != "tmi" && msg.Username != botNick {
		// Log out the message to the db
		go p.db.InsertLog(time.Now(), msg.Channel, msg.Username, msg.Message)

		if msg.Channel == "#"+botNick {
			p.parseHomeCmd(msg)
		} else {
			p.parseChannelCmd(msg)
		}
	}
}

// Handle home channel commands
func (p *Parser) parseHomeCmd(msg IRCMessage) {
	if msg.Message == "!join" {
		p.joinChannel(msg.Username)
	} else if msg.Message == "!leave" {
		p.leaveChannel(msg.Username)
	}
}

// Parse commands from regular channels
func (p *Parser) parseChannelCmd(msg IRCMessage) {
	chatChan, err := p.manager.GetChannel(msg.Channel[1:])
	if err != nil {
		fmt.Println(err)
		return
	}
	if reGreeting.FindString(msg.Message) != "" {
		chatChan.SendGreeting(msg.Username)
	} else if reFarewell.FindString(msg.Message) != "" {
		chatChan.SendFarewell(msg.Username)
	} else if msg.Message == "!gamerdeath" {
		chatChan.SendGamerdeath()
	}
}

// Add a new DB entry, join the IRC channel, add the channel to the manager
func (p *Parser) joinChannel(username string) {
	fmt.Println("JOIN -> " + username)
	homeChannel, err := p.manager.GetChannel(botNick)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Check if they have already registered
	if p.manager.IsRegistered(username) {
		homeChannel.SendRegisterError(username)
		return
	}

	id := getChannelID(apiClient, username)
	if id == "" {
		fmt.Println("ERROR: API Can't get ID for: " + username)
		return
	}

	go p.db.AddChannel(username, id)
	p.manager.RegisterChannel(username, id)
	homeChannel.SendRegistered(username)
}

// Remove the DB entry, leave the IRC channel, delete the channel from the manager
func (p *Parser) leaveChannel(username string) {
	fmt.Println("LEAVE -> " + username)
	homeChannel, err := p.manager.GetChannel(botNick)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Check if they have already unregistered
	if !p.manager.IsRegistered(username) {
		homeChannel.SendUnRegisterError(username)
		return
	}

	go p.db.DeleteChannelUser(username)
	p.manager.UnregisterChannel(username)
	homeChannel.SendUnregistered(username)
}
