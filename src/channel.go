package main

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	greetingCooldown   = 10
	farewellCooldown   = 10
	gamerdeathCooldown = 60
	registerCooldown   = 1
	reminderPeriod     = 10800
	offlineCheckRate   = 300
)

type ChatChannel struct {
	channelName       string
	id                string
	conn              *IrcConnection
	isGreetingReady   bool
	isFarewellReady   bool
	isGamerdeathReady bool
	isRegisterReady   bool
	timerStop         chan bool
}

var GreetingResponses = []string{
	"Hi",
	"Hello",
	"Hiya",
	"Hey",
	"Yo",
	"What's up",
	"How's it going",
	"Greetings",
	"Sup",
	"What's good",
	"Hey there",
	"Howdy",
	"Good to see you",
	"vvhat's up",
	"Henlo",
	"Hovvdy",
}

var FarewellResponses = []string{
	"Bye",
	"Goodnight",
	"Good night",
	"Goodbye",
	"Good bye",
	"See you",
	"See ya",
	"So long",
	"Farewell",
	"Later",
	"Seeya",
	"Ciao",
	"Au revoir",
	"Bon voyage",
	"Peace",
	"In a while crocodile",
	"See you later alligator",
	"Later alligator",
	"Have a good one",
	"l8r",
	"Later skater",
	"Catch you on the flip side",
	"Sayonara",
	"Auf weidersehen",
}

var SubGifters = []string{
	"technotoast",
	"kelleymcches",
	"wincerind",
	"hetero_corgi",
	"spoonlessalakazam",
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
		timerStop:         make(chan bool, 1),
	}
}

func (c *ChatChannel) SendGreeting(targetUser string) {
	if c.isGreetingReady {
		c.conn.Chat(c.channelName, getRandomGreeting(targetUser))
		go c.setGreetingTimer()
	}
}

func (c *ChatChannel) SendFarewell(targetUser string) {
	if c.isFarewellReady {
		c.conn.Chat(c.channelName, getRandomFarewell(targetUser))
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
	fmt.Println("Starting new getup timer thread for: ", c.channelName)
	for {
		select {
		case <-c.timerStop:
			fmt.Println("Stopping getup timer thread for: ", c.channelName)
			return
		default:
			uptime, err := getChannelUptime(apiClient, c.channelName)
			if err != nil {
				time.Sleep(time.Duration(60) * time.Second)
				continue
			}

			if uptime != -1 {
				waitTime := reminderPeriod - (uptime % reminderPeriod)
				fmt.Println("Waiting on tick for: ", c.channelName, " in: ", waitTime)
				// Timer ticks at the next 3 hour mark determined by uptime
				timer := time.NewTimer(time.Duration(waitTime) * time.Second)
				<-timer.C
				uptime, _ = getChannelUptime(apiClient, c.channelName)
				if c.conn.isConnected && uptime != -1 {
					select {
					case <-c.timerStop:
						fmt.Println("Stopping getup timer from inner for: ", c.channelName)
						return
					default:
						fmt.Println("TIMER TICK: ", c.channelName)
						c.conn.Chat(c.channelName, "MrDestructoid "+c.channelName+" alert! It's been 3 hours and its time to prevent Gamer Death!")
					}
				}
			} else {
				time.Sleep(time.Duration(offlineCheckRate) * time.Second)
			}
		}
	}
}

func (c *ChatChannel) StopGetupTimer() {
	c.timerStop <- true
}

func getRandomGreeting(targetUser string) string {
	n := rand.Int() % len(GreetingResponses)
	response := GreetingResponses[n] + " " + targetUser + " etalWave"

	if targetUser == "evanito" {
		response = response + " You are my favorite chatter :)"
	}

	for _, sub := range SubGifters {
		if sub == targetUser {
			response = response + " Thank you for the sub btw! :)"
			break
		}
	}
	return response
}

func getRandomFarewell(targetUser string) string {
	n := rand.Int() % len(FarewellResponses)
	return FarewellResponses[n] + " " + targetUser + " etalWave"
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
