package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"fmt"
	"time"
)

const (
	logTable      = "MonkaNetLog"
	channelsTable = "Channels"
)

type LogMessage struct {
	ChannelID string
	// Microseconds since epoch
	Timestamp   int64
	ChannelName string
	Message     string
	BadgeInfo   string
	Badges      string
	ClientNonce string
	Color       string
	DisplayName string
	Emotes      string
	MessageID   string
	Mod         string
	Reply       ReplyMessage
	UserID      string
	Username    string
}

type ReplyMessage struct {
	ReplyParentDisplayName string
	ReplyParentMessageBody string
	ReplyParentMessageID   string
	ReplyParentUserID      string
	ReplyParentUserLogin   string
}

type Channel struct {
	ChannelID   string
	ChannelName string
	TimeAdded   string
}

// DynamoConnection represents a connection state to the channel Database Server
type DynamoConnection struct {
	conn   *dynamodb.DynamoDB
	keyID  string
	secret string
}

// NewDynamoConnection returns a DynamoConnection object
func NewDynamoConnection(keyID string, secret string) *DynamoConnection {
	return &DynamoConnection{
		conn:   nil,
		keyID:  keyID,
		secret: secret,
	}
}

// Open initializes a new database connection
func (db *DynamoConnection) Open() error {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-2"),
		Credentials: credentials.NewStaticCredentials(db.keyID, db.secret, ""),
	})
	if err != nil {
		return err
	}
	db.conn = dynamodb.New(sess)
	return nil
}

// Close closes the database connection
func (db *DynamoConnection) Close() error {
	return nil
}

// AddChannel registers a new channel
func (db *DynamoConnection) AddChannel(username string, id string) {
	channel := Channel{
		ChannelID:   id,
		ChannelName: username,
		TimeAdded:   time.Now().UTC().Format(time.RFC850),
	}
	av, err := dynamodbattribute.MarshalMap(channel)
	if err != nil {
		fmt.Println("[DYNAMO] Error marshalling item:")
		fmt.Println(err.Error())
		return
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(channelsTable),
	}

	_, err = db.conn.PutItem(input)
	if err != nil {
		fmt.Println("[DYNAMO] Error calling PutItem:")
		fmt.Println(err.Error())
		return
	}
}

// DeleteChannelUser deletes a registered channel by username
func (db *DynamoConnection) DeleteChannelUser(username string) {
	// Remove this
}

// DeleteChannelID deletes a registered channel by id
func (db *DynamoConnection) DeleteChannelID(id string) {
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"ChannelID": {
				S: aws.String(id),
			},
		},
		TableName: aws.String(channelsTable),
	}

	_, err := db.conn.DeleteItem(input)
	if err != nil {
		fmt.Println("[DYNAMO] Error calling DeleteItem:")
		fmt.Println(err.Error())
		return
	}
}

// InsertLog inserts a log
func (db *DynamoConnection) InsertLog(msg *IRCMessage) {
	log := ircToLog(msg)
	av, err := dynamodbattribute.MarshalMap(log)
	if err != nil {
		fmt.Println("[DYNAMO] Error marshalling item:")
		fmt.Println(err.Error())
		return
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(logTable),
	}

	_, err = db.conn.PutItem(input)
	if err != nil {
		fmt.Println("[DYNAMO] Error calling PutItem:")
		fmt.Println(err.Error())
		fmt.Println(msg)
		return
	}
}

// GetRegisteredChannels returns two slices of channelnames and ids that are registered
func (db *DynamoConnection) GetRegisteredChannels() ([]string, []string) {
	result, err := db.conn.Scan(&dynamodb.ScanInput{
		TableName: aws.String(channelsTable),
	})
	if err != nil {
		fmt.Println(err.Error())
		return []string{}, []string{}
	}

	channels := []string{}
	ids := []string{}
	for _, i := range result.Items {
		item := Channel{}
		err = dynamodbattribute.UnmarshalMap(i, &item)
		if err != nil {
			fmt.Println("[DYNAMO] Got error unmarshalling:")
			fmt.Println(err.Error())
			return []string{}, []string{}
		}

		channels = append(channels, item.ChannelName)
		ids = append(ids, item.ChannelID)
	}
	return channels, ids
}

func ircToLog(msg *IRCMessage) *LogMessage {
	log := &LogMessage{
		ChannelID:   msg.ChannelID,
		Timestamp:   msg.Timestamp,
		ChannelName: msg.Channel[1:],
		Message:     msg.Message,
		Username:    msg.Username,
	}
	log.BadgeInfo = getTag(&msg.Tags, "badge-info")
	log.Badges = getTag(&msg.Tags, "badges")
	log.ClientNonce = getTag(&msg.Tags, "client-nonce")
	log.Color = getTag(&msg.Tags, "color")
	log.DisplayName = getTag(&msg.Tags, "display-name")
	log.Emotes = getTag(&msg.Tags, "emotes")
	log.MessageID = getTag(&msg.Tags, "id")
	log.Mod = getTag(&msg.Tags, "mod")
	log.Reply = ReplyMessage{
		ReplyParentDisplayName: getTag(&msg.Tags, "reply-parent-display-name"),
		ReplyParentMessageBody: getTag(&msg.Tags, "reply-parent-msg-body"),
		ReplyParentMessageID:   getTag(&msg.Tags, "reply-parent-msg-id"),
		ReplyParentUserID:      getTag(&msg.Tags, "reply-parent-user-id"),
		ReplyParentUserLogin:   getTag(&msg.Tags, "reply-parent-user-login"),
	}
	log.UserID = getTag(&msg.Tags, "user-id")
	return log
}

func getTag(tags *IRCTags, tagName string) string {
	if data, ok := (*tags)[tagName]; ok {
		return data
	}
	return ""
}
