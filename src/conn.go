package main

import (
	"fmt"
	"net"
	"strings"
)

const (
	ircHostURL  = "irc.twitch.tv"
	ircHostPort = "6667"
)

type IrcConnection struct {
	conn        net.Conn
	isConnected bool
	connList    []string
}

// Returns a new IRC Client
func NewIRCConnection(conns []string) *IrcConnection {
	return &IrcConnection{
		conn:        nil,
		isConnected: false,
		connList:    conns,
	}
}

// Connect to the IRC server, authenticate and join target channels
func (c *IrcConnection) Connect(nick string, pass string) error {
	if !c.isConnected {
		conn, err := net.Dial("tcp", ircHostURL+":"+ircHostPort)
		if err != nil {
			fmt.Println(err)
			return err
		}
		c.conn = conn
		c.send("PASS " + pass)
		c.send("NICK " + nick)
		for _, channel := range c.connList {
			c.send("JOIN #" + channel)
			fmt.Println("Joining: " + channel)
		}
		c.isConnected = true
	}
	return nil
}

// Disconnect from the IRC server
func (c *IrcConnection) Disconnect() error {
	if c.isConnected {
		err := c.conn.Close()
		if err != nil {
			fmt.Println(err)
			return err
		}
		c.isConnected = false
	}
	return nil
}

// Send a chat message to an IRC channel
func (c *IrcConnection) Chat(channel string, message string) {
	c.send("PRIVMSG #" + channel + " :" + message)
}

// Receieve data from the IRC connection. Handles ping pong automatically.
func (c *IrcConnection) Recv() (string, error) {
	buf := make([]byte, 4096)
	len, err := c.conn.Read(buf)
	if err != nil {
		c.Disconnect()
		return "", err
	}
	// Cast to string, trim newlines
	message := strings.TrimSpace(string(buf[:len]))

	// Handle ping pong, return the next recv
	if message == "PING :tmi.twitch.tv" {
		c.send("PONG :tmi.twitch.tv")
		message, err = c.Recv()
		if err != nil {
			return "", err
		}
	}

	//fmt.Println("RX: " + message)
	return message, nil
}

// Raw send to the socket
func (c *IrcConnection) send(message string) {
	fmt.Fprintf(c.conn, message+"\r\n")
	fmt.Println("TX: " + message)
}
