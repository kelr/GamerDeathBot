package main

import (
	"time"
)

const (
	greetingCooldown   = 10
	farewellCooldown   = 10
	gamerdeathCooldown = 60
)

type ChatChannel struct {
	channelName       string
	id                string
	conn              *IrcConnection
	isGreetingReady   bool
	isFarewellReady   bool
	isGamerdeathReady bool
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
