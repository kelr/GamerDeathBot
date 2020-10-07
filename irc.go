package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	rxBufSize   = 4096
	txQueueSize = 100
	pingMessage = "PING :tmi.twitch.tv"
	pongMessage = "PONG :tmi.twitch.tv"
	rateLimit = 2 * time.Second
)

// IrcConnection represents a connection state to an IRC server over a TCP socket
type IrcConnection struct {
	host        string
	port        string
	conn        net.Conn
	isConnected bool
	txQueue     chan string
	control     chan bool
}

// NewIRCConnection returns a new IRC Client
func NewIRCConnection(host string, port string) *IrcConnection {
	return &IrcConnection{
		host:        host,
		port:        port,
		conn:        nil,
		isConnected: false,
		txQueue:     make(chan string, txQueueSize),
		control:     make(chan bool),
	}
}

// Connect to the IRC server, authenticate and join target channels
func (c *IrcConnection) Connect(nick string, pass string) error {
	if !c.isConnected {
		conn, err := net.Dial("tcp", c.host+":"+c.port)
		if err != nil {
			fmt.Println(err)
			return err
		}
		c.conn = conn
		go c.rateLimiter()
		c.authenticate(nick, pass)
		c.isConnected = true
	}
	return nil
}

func (c *IrcConnection) authenticate(nick string, pass string) {
	c.send("PASS " + pass)
	c.send("NICK " + nick)
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

// Join sends a JOIN command to the IRC server to join a channel
func (c *IrcConnection) Join(channel string) {
	c.send("JOIN #" + channel)
}

// Part sends a PART command to the IRC server to leave a channel
func (c *IrcConnection) Part(channel string) {
	c.send("PART #" + channel)
}

// Chat sends a PRIVMSG to an IRC channel
func (c *IrcConnection) Chat(channel string, message string) {
	c.txQueue <- "PRIVMSG #" + channel + " :" + message
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
		time.Sleep(rateLimit)
	}
}

// Recv receieve data from the IRC connection. Handles ping pong automatically.
func (c *IrcConnection) Recv() (string, error) {
	buf := make([]byte, rxBufSize)
	len, err := c.conn.Read(buf)
	if err != nil {
		fmt.Println(err)
		c.Disconnect()
		return "", err
	}
	// Cast to string, trim newlines
	message := strings.TrimSpace(string(buf[:len]))
	message, err = c.handlePingPong(message)
	if err != nil {
		return "", err
	}

	return message, nil
}

// Handle ping pong, return the next recv instead
func (c *IrcConnection) handlePingPong(message string) (string, error) {
	if message == pingMessage {
		c.send(pongMessage)
		message, err := c.Recv()
		if err != nil {
			return "", err
		}
		return message, nil
	}
	return message, nil
}

// Raw send to the socket
func (c *IrcConnection) send(message string) error {
	_, err := fmt.Fprintf(c.conn, message+"\r\n")
	if err != nil {
		fmt.Println(err)
		c.Disconnect()
		return err
	}
	fmt.Println("TX: " + message)
	return nil
}
