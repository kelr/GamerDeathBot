#!/usr/bin/python
"""Main Module"""

import re

from conn import SocketConnection
from api.client import TwitchAPIClient
from channel import ChannelTransmit
import consts

REGEX_MESSAGE = re.compile(r"^:\w+!\w+@\w+\.tmi\.twitch\.tv PRIVMSG #\w+ :")

REGEX_GREETING = r"(hi|hello|hey|yo|sup|howdy|hovvdy|greetings|what's good|whats good|vvhat's good|vvhats good|what's up|whats up|vvhat's up|vvhats up) @*GamerDeathBot"

def parse_msg(conn, msg, active_channels):
    """Thread to read from the socket connection.

    Args:
        conn -- SocketConnection object
        msg -- message to parse
        active_channels -- active channel transmit dictionary
    """
    # Handle ping pong, once ~5 mins
    if msg == "PING :tmi.twitch.tv":
        conn.send("PONG :tmi.twitch.tv")
        return

    username, message, channel = split_msg_data(msg)

    print("RX: " + str(channel) + " -- " + username + ": " + message.strip())

    # Match a greeting message
    if re.match(REGEX_GREETING, message, re.IGNORECASE):
        active_channels[channel].send_greeting(conn, channel, username)

    # Match a gamerdeath message
    elif message == "!gamerdeath":
        active_channels[channel].send_gamerdeath(conn, channel)

def split_msg_data(msg):
    """Split out user, message and channel names from a receieved message

    Args:
        msg -- message to parse

    Return:
        username, message, channel 3-tuple. Strings, but channel can be None.
    """
    username = re.search(r"\w+", msg).group(0)
    message = REGEX_MESSAGE.sub("", msg)
    channel = re.search(r"#\w+", msg)
    if channel:
        channel = channel.group(0)
    return username, message, channel

def main():
    """Setup the socket connection and the rx thread."""
    api = TwitchAPIClient(consts.CLIENT_ID, consts.PASS)
    conn = SocketConnection()

    active_channels = {}

    for chan in consts.TARGET_CHANNELS:
        active_channels[chan] = ChannelTransmit(conn, api, chan)

    conn.connect()

    # Rx and process messages forever
    while True:
        response = conn.recv()
        if response is not None:
            parse_msg(conn, response, active_channels)

if __name__ == '__main__':
    main()
