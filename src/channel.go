package main

import (
	"time"
)

const (
	greetingCooldown   = 10
	farewellCooldown   = 10
	gamerdeathCooldown = 60
	registerCooldown   = 1
	reminderPeriod     = 10800
)

type ChatChannel struct {
	channelName       string
	id                string
	conn              *IrcConnection
	isGreetingReady   bool
	isFarewellReady   bool
	isGamerdeathReady bool
	isRegisterReady   bool
}

// Returns a new IRC Client
func NewChatChannel(username string, channelId string, connection *IrcConnection) *ChatChannel {
	return &ChatChannel{
		channelName:       username,
		id:                channelId,
		conn:              connection,
		isGreetingReady:   true,
		isFarewellReady:   true,
		isGamerdeathReady: true,
		isRegisterReady:   true,
	}
}

func (c *ChatChannel) SendGreeting(targetUser string) {
	if c.isGreetingReady {
		c.conn.Chat(c.channelName, getRandomGreeting())
		go c.setGreetingTimer()
	}
}

func (c *ChatChannel) SendFarewell(targetUser string) {
	if c.isFarewellReady {
		c.conn.Chat(c.channelName, getRandomFarewell())
		go c.setFarewellTimer()
	}
}

func (c *ChatChannel) SendGamerdeath() {
	if c.isGamerdeathReady {
		c.conn.Chat(c.channelName, "MrDestructoid Chat, remember to get up and stretch to prevent Gamer Death!")
		go c.setGamerdeathTimer()
	}
}

func (c *ChatChannel) SendRegistered(targetUser string) {
	if c.isRegisterReady {
		c.conn.Chat(c.channelName, "I joined your chat, "+targetUser+"!")
		go c.setRegisterCooldown()
	}
}

func (c *ChatChannel) SendUnregistered(targetUser string) {
	if c.isRegisterReady {
		c.conn.Chat(c.channelName, "I left your chat, "+targetUser+"!")
		go c.setRegisterCooldown()
	}
}

func (c *ChatChannel) SendRegisterError(targetUser string) {
	if c.isRegisterReady {
		c.conn.Chat(c.channelName, "I'm already in your chat, "+targetUser+"!")
		go c.setRegisterCooldown()
	}
}

func (c *ChatChannel) SendUnRegisterError(targetUser string) {
	if c.isRegisterReady {
		c.conn.Chat(c.channelName, "I've already left your chat, "+targetUser+"!")
		go c.setRegisterCooldown()
	}
}

func (c *ChatChannel) StartGetupTimer() {
	for {
		uptime := getChannelUptime(c.channelName)

		if uptime != -1 {
			// Timer ticks at the next 3 hour mark determined by uptime
			timer := time.NewTimer(time.Duration(reminderPeriod-(uptime%reminderPeriod)) * time.Second)
			<-timer.C
			if c.conn.isConnected {
				c.conn.Chat(c.channelName, "MrDestructoid "+c.channelName+" alert! It's been 3 hours and its time to prevent Gamer Death!")
			}
		} else {
			time.Sleep(5 * time.Second)
		}
	}
}

// TODO lists
func getRandomGreeting() string {
	return "hi"
}

// TODO lists
func getRandomFarewell() string {
	return "bye"
}

func (c *ChatChannel) setGreetingTimer() {
	c.isGreetingReady = false
	timer := time.NewTimer(greetingCooldown * time.Second)
	<-timer.C
	c.isGreetingReady = true
}

func (c *ChatChannel) setFarewellTimer() {
	c.isFarewellReady = false
	timer := time.NewTimer(farewellCooldown * time.Second)
	<-timer.C
	c.isFarewellReady = true
}

func (c *ChatChannel) setGamerdeathTimer() {
	c.isGamerdeathReady = false
	timer := time.NewTimer(gamerdeathCooldown * time.Second)
	<-timer.C
	c.isGamerdeathReady = true
}

func (c *ChatChannel) setRegisterCooldown() {
	c.isRegisterReady = false
	timer := time.NewTimer(registerCooldown * time.Second)
	<-timer.C
	c.isRegisterReady = true
}
