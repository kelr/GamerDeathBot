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

	opt := &twitchapi.GetStreamsOpt{
		UserLogin: username,
	}

	response, err := client.GetStreams(opt)
	if err != nil {
		fmt.Println("Error in API call:", err)
	}

	if len(response.Data) > 0 {
		t, _ := time.Parse(time.RFC3339, response.Data[0].StartedAt)
		return int(time.Since(t).Seconds())
	}
	return -1
}

func getChannelID(username string) string {

	client := twitchapi.NewTwitchClient(clientID)

	opt := &twitchapi.GetUsersOpt{
		Login: username,
	}

	response, err := client.GetUsers(opt)
	if err != nil {
		fmt.Println("Error in API call:", err)
	}

	if len(response.Data) > 0 {
		return response.Data[0].ID
	}
	return ""
}
