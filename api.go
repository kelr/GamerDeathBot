package main

import (
	"fmt"
	"time"

	"github.com/kelr/gundyr/helix"
)

// API is an interface to the underlying Helix API connection
type API interface {
	GetChannelUptime(username string) (int, error)
	GetChannelID(username string) (string, error)
}

// APIClient wraps a Twitch Helix API Client
type APIClient struct {
	api *helix.Client
}

// NewAPIClient returns a new API Client
func NewAPIClient(id string, secret string) (*APIClient, error) {
	config := &helix.Config{
		ClientID:     id,
		ClientSecret: secret,
	}
	api, err := helix.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &APIClient{
		api: api,
	}, nil
}

// GetChannelUptime returns channel uptime in an integer number of seconds or -1 if not live
func (a *APIClient) GetChannelUptime(username string) (int, error) {
	opt := &helix.GetStreamsOpt{
		UserLogin: username,
	}

	response, err := a.api.GetStreams(opt)
	if err != nil {
		fmt.Println("Error in API call:", err)
		return -1, err
	}

	if len(response.Data) > 0 {
		t, _ := time.Parse(time.RFC3339, response.Data[0].StartedAt)
		return int(time.Since(t).Seconds()), nil
	}
	return -1, nil
}

// GetChannelID returns the ID of a Twitch User
func (a *APIClient) GetChannelID(username string) (string, error) {
	opt := &helix.GetUsersOpt{
		Login: []string{username},
	}

	response, err := a.api.GetUsers(opt)
	if err != nil {
		fmt.Println("Error in API call:", err)
		return "", err
	}

	if len(response.Data) > 0 {
		return response.Data[0].ID, nil
	}
	return "", err
}
