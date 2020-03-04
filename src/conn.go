package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	ircHostURL  = "irc.twitch.tv"
	ircHostPort = "6667"
)

type IrcConnection struct {
	conn        net.Conn
	isConnected bool
	connList    []string
	txQueue chan string
	control chan bool
}

// Returns a new IRC Client
func NewIRCConnection(conns []string) *IrcConnection {
	return &IrcConnection{
		conn:        nil,
		isConnected: false,
		connList:    conns,
		txQueue: make(chan string, 100),
		control: make(chan bool),
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
		go c.rateLimiter()
		c.send("PASS " + pass)
		c.send("NICK " + nick)
		for _, channel := range c.connList {
			c.Join(channel)
		}
		c.isConnected = true
	}
	return nil
}

func (c *IrcConnection) Join(channel string) {
	c.send("JOIN #" + channel)
}

func (c *IrcConnection) Part(channel string) {
	c.send("PART #" + channel)
}

// Disconnect from the IRC server
func (c *IrcConnection) Disconnect() error {
	if c.isConnected {
		c.control <- true
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
	c.txQueue <- "PRIVMSG #" + channel + " : " + message

}

// Rate limit the transmission of messages to the IRC server
func (c *IrcConnection) rateLimiter() {
	for {
		select {
        case <-c.control:
            return
        case msg := <-c.txQueue:
            c.send(msg)
        }
        time.Sleep(2 * time.Second)
	}
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
	fmt.Println("RX: " + message)
	return message, nil
}

// Raw send to the socket
func (c *IrcConnection) send(message string) {
	fmt.Fprintf(c.conn, message+"\r\n")
	fmt.Println("TX: " + message)
}
