package main

import (
	"errors"
)

// ChannelManager handles all the registered channels for this bot instance.
type ChannelManager struct {
	channels map[string]*ChatChannel
}

// NewChannelManager returns a new Channel Manager
func NewChannelManager(userList []string, idList []string, irc *IrcConnection) *ChannelManager {
	c := &ChannelManager{
		channels: make(map[string]*ChatChannel),
	}
	for index, channel := range userList {
		c.RegisterChannel(channel, idList[index], irc)
	}
	return c
}

// RegisterChannel registers a new channel if it is not already registered.
func (c *ChannelManager) RegisterChannel(channel string, id string, irc *IrcConnection) {
	if !c.IsRegistered(channel) {
		c.channels[channel] = NewChatChannel(channel, id, irc)
		go c.channels[channel].StartGetupTimer()
	}
}

// UnregisterChannel unregisters a channel if it exists.
func (c *ChannelManager) UnregisterChannel(channel string) {
	if c.IsRegistered(channel) {
		c.channels[channel].StopGetupTimer()
		delete(c.channels, channel)
	}
}

// StartAllTimers starts all registered timers
func (c *ChannelManager) StartAllTimers() {
	for channel := range c.channels {
		go c.channels[channel].StartGetupTimer()
	}
}

// GetChannel retrieves a registered channel
func (c *ChannelManager) GetChannel(channel string) (*ChatChannel, error) {
	if c.IsRegistered(channel) {
		return c.channels[channel], nil
	}
	return nil, errors.New("Channel " + channel + " has not been registered.")
}

// IsRegistered check if a channel is already registered or not
func (c *ChannelManager) IsRegistered(channel string) bool {
	if _, ok := c.channels[channel]; ok {
		return true
	}
	return false
}
