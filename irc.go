package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"
)

const (
	rxBufSize     = 4096
	txQueueSize   = 100
	pingMessage   = "PING :tmi.twitch.tv"
	pongMessage   = "PONG :tmi.twitch.tv"
	rateLimit     = 2 * time.Second
	regexUsername = `\w+`
	regexChannel  = `#\w+`
	regexMessage  = `^:\w+!\w+@\w+\.tmi\.twitch\.tv PRIVMSG #\w+ :`
)

var (
	reUser    = regexp.MustCompile(regexUsername)
	reChannel = regexp.MustCompile(regexChannel)
	reMessage = regexp.MustCompile(regexMessage)
)

// IRC is an interface for the underlying IRC connection
type IRC interface {
	Connect(login string, token string) error
	Disconnect() error
	Join(channel string)
	Part(channel string)
	Chat(channel string, message string)
	Read() (*IRCMessage, error)
	IsConnected() bool
}

// IRCMessage represents parsed out fields of a message from IRC
type IRCMessage struct {
	Channel     string
	Username    string
	Message     string
	Command     string
	CommandArgs []string
	Tags        IRCTags
	Timestamp   time.Time
}

// IRCTags represents a string:string map of IRC Tags
type IRCTags map[string]string

// IrcConnection represents a connection state to an IRC server over a TCP socket
type IrcConnection struct {
	host           string
	port           string
	login          string
	token          string
	conn           net.Conn
	isConnected    bool
	txQueue        chan string
	control        chan bool
	reader         *bufio.Reader
	connectionList map[string]bool
}

// NewIRCConnection returns a new IRC Client.
func NewIRCConnection(host string, port string) *IrcConnection {
	return &IrcConnection{
		host:           host,
		port:           port,
		login:          "",
		token:          "",
		conn:           nil,
		isConnected:    false,
		txQueue:        make(chan string, txQueueSize),
		control:        make(chan bool),
		connectionList: make(map[string]bool),
	}
}

// Connect to the IRC server and
// Login is the login username for the account and token is an
// OAuth2 token with Twitch IRC permissions, prefixed with oauth:
func (c *IrcConnection) Connect(login string, token string) error {
	if login == "" || token == "" {
		return errors.New("[IRC]: cannot connect, missing login or OAuth token")
	}
	if !c.isConnected {
		conn, err := net.Dial("tcp", c.host+":"+c.port)
		if err != nil {
			fmt.Println(err)
			return err
		}
		c.conn = conn
		c.reader = bufio.NewReader(c.conn)

		go c.rateLimiter()
		if err = c.authenticate(login, token); err != nil {
			return err
		}
	}
	return nil
}

// Attempt to authenticate with the IRC server, and read the response
func (c *IrcConnection) authenticate(login string, token string) error {
	c.send("PASS " + token)
	c.send("NICK " + login)

	// Initial read to parse welcome message and confirm authentication
	if _, err := c.Read(); err != nil {
		return err
	}
	return nil
}

// IsConnected returns the connection state
func (c *IrcConnection) IsConnected() bool {
	return c.isConnected
}

func (c *IrcConnection) enableCaps() {
	c.send("CAP REQ :twitch.tv/tags twitch.tv/commands")
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

// Join sends a JOIN command to the IRC server to join a channel
// Adds the channel to the connection list if it is not already there.
func (c *IrcConnection) Join(channel string) {
	if _, ok := c.connectionList[channel]; !ok {
		c.connectionList[channel] = true
	}
	c.send("JOIN #" + channel)
}

// Part sends a PART command to the IRC server to leave a channel
// Removes the channel from the connection list if it is there.
func (c *IrcConnection) Part(channel string) {
	if _, ok := c.connectionList[channel]; ok {
		delete(c.connectionList, channel)
	}
	c.send("PART #" + channel)
}

// Chat sends a PRIVMSG to an IRC channel
func (c *IrcConnection) Chat(channel string, message string) {
	c.txQueue <- "PRIVMSG #" + channel + " :" + message
}

// Read receieve data from the IRC connection. Handles ping pong automatically.
func (c *IrcConnection) Read() (*IRCMessage, error) {
	message, err := c.reader.ReadString('\n')
	if err != nil {
		c.Disconnect()
		return nil, err
	}

	// Parse out the message
	parsedMsg, err := parseIRCMessage(message)
	if err != nil {
		return nil, err
	}

	c.respondDefaultCmds(parsedMsg)

	return parsedMsg, nil
}

// Raw send to the socket
func (c *IrcConnection) send(message string) error {
	_, err := fmt.Fprintf(c.conn, message+"\r\n")
	if err != nil {
		fmt.Println(err)
		c.Disconnect()
		return err
	}
	return nil
}

func (c *IrcConnection) respondDefaultCmds(msg *IRCMessage) {
	if msg.Command == "001" {
		c.handle001()
	} else if msg.Command == "PING" {
		c.handlePing()
	}
}

// Handle ping pong, return the next read instead
func (c *IrcConnection) handlePing() {
	c.send(pongMessage)
}

func (c *IrcConnection) handle001() {
	c.enableCaps()
	c.isConnected = true
	// Join channels in the connection list. This handles re-joining channels
	// if we had to re-establish communication with the IRC server.
	for channel := range c.connectionList {
		c.Join(channel)
	}
}

// Parse parses out commands from a chat message into an IRCMessage object
func parseIRCMessage(msg string) (*IRCMessage, error) {
	message := strings.TrimRight(msg, "\r\n")
	if len(message) == 0 {
		return nil, errors.New("IRC Error: Cannot parse empty message")
	}

	ircMessage := &IRCMessage{
		Tags:      IRCTags{},
		Timestamp: time.Now().UTC(),
	}

	// Parse out tags if they exist
	if strings.HasPrefix(message, "@") {
		tagEnd := strings.Index(message, " ")
		if tagEnd == -1 {
			return nil, errors.New("IRC Error: Parsing failed, missing data after tag: " + message)
		}
		ircMessage.Tags = parseTags(message[:tagEnd])
		message = message[tagEnd+1:]
	}

	// Parse out the message header
	if strings.HasPrefix(message, ":") {
		headerEnd := strings.Index(message, " ")
		if headerEnd == -1 {
			return nil, errors.New("IRC Error: Parsing failed, missing data after header: " + message)
		}
		ircMessage.Username = parseUsername(message[:headerEnd])
		message = message[headerEnd+1:]
	}

	// Parse out the rest of the message
	split := strings.SplitN(message, " :", 2)
	if split[0] == "" {
		return nil, errors.New("IRC Error: Parsing failed, missing IRC Command: " + message)
	}
	ircMessage.Message = split[1]

	// Parse out the command and any args
	commands := strings.Split(split[0], " ")
	ircMessage.Command = commands[0]
	ircMessage.Channel = parseChannel(commands)
	ircMessage.CommandArgs = commands[1:]

	return ircMessage, nil
}

// Parse out the channel if it exists
func parseChannel(commands []string) string {
	for i := 0; i < len(commands); i++ {
		if strings.HasPrefix(commands[i], "#") {
			return commands[i]
		}
	}
	return ""
}

// Split tags by semicolon and add them into a map[string]string
func parseTags(msg string) IRCTags {
	ret := IRCTags{}
	if !strings.HasPrefix(msg, "@") {
		return ret
	}
	// Strip off the prefix
	msg = msg[1:]

	tags := strings.Split(msg, ";")
	for _, tag := range tags {
		parts := strings.SplitN(tag, "=", 2)
		if len(parts) < 2 {
			ret[parts[0]] = ""
		} else {
			ret[parts[0]] = parts[1]
		}
	}
	return ret
}

func parseUsername(msg string) string {
	if !strings.HasPrefix(msg, ":") {
		return ""
	}
	// Strip off the prefix
	msg = msg[1:]

	usernameEnd := strings.Index(msg, "!")
	if usernameEnd != -1 {
		return msg[:usernameEnd]
	}

	usernameEnd = strings.Index(msg, "@")
	if usernameEnd != -1 {
		return msg[:usernameEnd]
	}

	return msg
}

func prettyPrintIRC(msg IRCMessage) {
	fmt.Println("Timestamp:", msg.Timestamp)
	fmt.Println("Channel:", msg.Channel)
	fmt.Println("Command:", msg.Command)
	fmt.Println("CommandArgs:", msg.CommandArgs)
	fmt.Println("Username:", msg.Username)
	fmt.Println("Message:", msg.Message)
	fmt.Println("Tags:", msg.Tags)
	fmt.Println()
}
