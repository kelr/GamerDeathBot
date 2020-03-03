package main

import (
	//"encoding/json"
	"fmt"
	"github.com/kelr/go-twitch-api/twitchapi"
	"time"
)

// Returns channel uptime in an integer number of seconds or -1 if not live
func getChannelUptime(username string) int {

	client := twitchapi.NewTwitchClient(clientID)

	// Set options, English and only return the top 2 streams
	opt := &twitchapi.GetStreamsOpt{
		UserLogin: username,
	}

	// Returns a GetStreamsResponse object
	response, err := client.GetStreams(opt)
	if err != nil {
		fmt.Println("Error:", err)
	}

	if len(response.Data) > 0 {
		t, _ := time.Parse(time.RFC3339, response.Data[0].StartedAt)
		return int(time.Since(t).Seconds())
	}
	return -1

}

// TODO
func getChannelID(username string) string {

	return "31903323"

}
