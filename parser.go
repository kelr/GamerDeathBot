package main

import (
	"fmt"
	"regexp"
	"time"
)

const (
	regexGreeting = `(?i)(hi|hiya|hello|hey|yo|sup|howdy|hovvdy|greetings|what's good|whats good|vvhat's good|vvhats good|what's up|whats up|vvhat's up|vvhats up|konichiwa|hewwo|etalWave|vvhats crackalackin|whats crackalackin|henlo|good morning|good evening|good afternoon) (@*GamerDeathBot|gdb)`
	regexFarewell = `(?i)(bye|goodnight|good night|goodbye|good bye|see you|see ya|so long|farewell|later|seeya|ciao|au revoir|bon voyage|peace|in a while crocodile|see you later alligator|later alligator|have a good one|igottago|l8r|later skater|catch you on the flip side|bye-bye|sayonara) (@*GamerDeathBot|gdb)`
)

var (
	reGreeting = regexp.MustCompile(regexGreeting)
	reFarewell = regexp.MustCompile(regexFarewell)
)

// Dispatcher determines what to execute per chat message
type Dispatcher struct {
	db      Database
	manager *ChannelManager
}

// NewDispatcher returns a new Dispatcher
func NewDispatcher(db Database, manager *ChannelManager) *Dispatcher {
	return &Dispatcher{
		db:      db,
		manager: manager,
	}
}

// Dispatch determines what command to run based on the input IRC Message
func (p *Dispatcher) Dispatch(msg *IRCMessage) {
	// Don't reply to self
	if msg.Command == "PRIVMSG" && msg.Username != botNick {
		// Log out the message to the db
		go p.db.InsertLog(time.Now(), msg.Channel, msg.Username, msg.Message)
		if msg.Channel == "#"+botNick {
			p.parseHomeCmd(msg)
		} else {
			p.parseChannelCmd(msg)
		}
	}
	fmt.Println(msg.Command, msg.Channel, msg.Username, msg.Message)
}

// Handle home channel commands
func (p *Dispatcher) parseHomeCmd(msg *IRCMessage) {
	if msg.Message == "!join" {
		p.joinChannel(msg.Username)
	} else if msg.Message == "!leave" {
		p.leaveChannel(msg.Username)
	}
}

// Parse commands from regular channels
func (p *Dispatcher) parseChannelCmd(msg *IRCMessage) {
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
func (p *Dispatcher) joinChannel(username string) {
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
	p.manager.RegisterChannel(username)
	homeChannel.SendRegistered(username)
}

// Remove the DB entry, leave the IRC channel, delete the channel from the manager
func (p *Dispatcher) leaveChannel(username string) {
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
