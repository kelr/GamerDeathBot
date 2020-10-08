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

type ChatChannel struct {
	channelName       string
	irc               IRC
	isGreetingReady   bool
	isFarewellReady   bool
	isGamerdeathReady bool
	isRegisterReady   bool
	timerStop         chan bool
	api               API
}

// Returns a new IRC Client
func NewChatChannel(username string, connection IRC, api API) *ChatChannel {
	return &ChatChannel{
		channelName:       username,
		irc:               connection,
		isGreetingReady:   true,
		isFarewellReady:   true,
		isGamerdeathReady: true,
		isRegisterReady:   true,
		timerStop:         make(chan bool, 1),
		api:               api,
	}
}

func (c *ChatChannel) JoinChannel() {
	c.irc.Join(c.channelName)
}

func (c *ChatChannel) LeaveChannel() {
	c.irc.Part(c.channelName)
}

func (c *ChatChannel) SendGreeting(targetUser string) {
	if c.isGreetingReady {
		c.irc.Chat(c.channelName, getRandomGreeting(targetUser))
		go c.setGreetingTimer()
	}
}

func (c *ChatChannel) SendFarewell(targetUser string) {
	if c.isFarewellReady {
		c.irc.Chat(c.channelName, getRandomFarewell(targetUser))
		go c.setFarewellTimer()
	}
}

func (c *ChatChannel) SendGamerdeath() {
	if c.isGamerdeathReady {
		c.irc.Chat(c.channelName, "MrDestructoid Chat, remember to get up and stretch to prevent Gamer Death!")
		go c.setGamerdeathTimer()
	}
}

func (c *ChatChannel) SendRegistered(targetUser string) {
	if c.isRegisterReady {
		c.irc.Chat(c.channelName, "I joined your chat, "+targetUser+"!")
		go c.setRegisterCooldown()
	}
}

func (c *ChatChannel) SendUnregistered(targetUser string) {
	if c.isRegisterReady {
		c.irc.Chat(c.channelName, "I left your chat, "+targetUser+"!")
		go c.setRegisterCooldown()
	}
}

func (c *ChatChannel) SendRegisterError(targetUser string) {
	if c.isRegisterReady {
		c.irc.Chat(c.channelName, "I'm already in your chat, "+targetUser+"!")
		go c.setRegisterCooldown()
	}
}

func (c *ChatChannel) SendUnRegisterError(targetUser string) {
	if c.isRegisterReady {
		c.irc.Chat(c.channelName, "I've already left your chat, "+targetUser+"!")
		go c.setRegisterCooldown()
	}
}

func (c *ChatChannel) StartGetupTimer() {
	for {
		select {
		case <-c.timerStop:
			fmt.Println("[GETUP]: Stopping getup timer thread for: ", c.channelName)
			return
		default:
			uptime, err := c.api.GetChannelUptime(c.channelName)
			if err != nil {
				time.Sleep(time.Duration(60) * time.Second)
				continue
			}

			if uptime != -1 {
				waitTime := reminderPeriod - (uptime % reminderPeriod)
				fmt.Println("[GETUP]: Waiting on tick for:", c.channelName, "in:", waitTime, "seconds")
				// Timer ticks at the next 3 hour mark determined by uptime
				timer := time.NewTimer(time.Duration(waitTime) * time.Second)
				<-timer.C
				uptime, _ = c.api.GetChannelUptime(c.channelName)
				if c.irc.IsConnected() && uptime != -1 {
					select {
					case <-c.timerStop:
						fmt.Println("[GETUP]: Stopping getup timer from inner for: ", c.channelName)
						return
					default:
						fmt.Println("[GETUP]: TIMER TICK: ", c.channelName)
						c.irc.Chat(c.channelName, "MrDestructoid "+c.channelName+" alert! It's been 3 hours and its time to prevent Gamer Death!")
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
