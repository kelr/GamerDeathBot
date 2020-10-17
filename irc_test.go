package main

import (
	"reflect"
	"testing"
)

func TestParseIRCMessage(t *testing.T) {
	t.Parallel()

	// Happy Path
	inputs := []struct {
		Input    string
		Expected IRCMessage
	}{
		{
			Input: ":peepo PRIVMSG #channel :normal message",
			Expected: IRCMessage{
				Channel:     "#channel",
				Username:    "peepo",
				Message:     "normal message",
				Command:     "PRIVMSG",
				CommandArgs: []string{"#channel"},
				Tags:        IRCTags{},
			},
		},
		{
			Input: "@tag1=one;tag2=222;tag3= :peepo!peepo@peepo PRIVMSG #channel :normal message",
			Expected: IRCMessage{
				Channel:     "#channel",
				Username:    "peepo",
				Message:     "normal message",
				Command:     "PRIVMSG",
				CommandArgs: []string{"#channel"},
				Tags: IRCTags{
					"tag1": "one",
					"tag2": "222",
					"tag3": "",
				},
			},
		},
		{
			Input: ":peepo PING :pong",
			Expected: IRCMessage{
				Channel:     "",
				Username:    "peepo",
				Message:     "pong",
				Command:     "PING",
				CommandArgs: []string{},
				Tags:        IRCTags{},
			},
		},
	}

	for _, test := range inputs {
		msg, err := parseIRCMessage(test.Input)
		if err != nil {
			t.Errorf("Input: %s, Expected a nil error", test.Input)
		}

		if msg == nil {
			t.Errorf("Input: %s, Expected a non message", test.Input)
		}

		if msg.Channel != test.Expected.Channel {
			t.Errorf("Input: %s, Expected: %s", msg.Channel, test.Expected.Channel)
		}
		if msg.Username != test.Expected.Username {
			t.Errorf("Input: %s, Expected: %s", msg.Username, test.Expected.Username)
		}
		if msg.Message != test.Expected.Message {
			t.Errorf("Input: %s, Expected: %s", msg.Message, test.Expected.Message)
		}
		if msg.Command != test.Expected.Command {
			t.Errorf("Input: %s, Expected: %s", msg.Command, test.Expected.Command)
		}
		if !reflect.DeepEqual(msg.CommandArgs, test.Expected.CommandArgs) {
			t.Errorf("Input: %s, Expected: %s", msg.CommandArgs, test.Expected.CommandArgs)
		}
		if !reflect.DeepEqual(msg.Tags, test.Expected.Tags) {
			t.Errorf("Input: %v, Expected: %v", msg.Tags, test.Expected.Tags)
		}
	}
}

func TestParseIRCMessageErrors(t *testing.T) {
	t.Parallel()

	// Errors
	inputs := []struct {
		Input string
	}{
		{""},
		{"@tags"},
		{":header"},
		{"@tags:header"},
		{" :"},
	}

	for _, test := range inputs {
		msg, err := parseIRCMessage(test.Input)
		if err == nil {
			t.Errorf("Input: %s, Expected a non nil error", test.Input)
		}

		if msg != nil {
			t.Errorf("Input: %s, Expected a nil message for parsed message with errors", test.Input)
		}
	}
}
