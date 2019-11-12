#!/usr/bin/python
"""Main Module"""

import re
import signal
import sys
import logging
from datetime import datetime

from conn import SocketConnection
from api.client import TwitchAPIClient
from channel import ChannelTransmit
import consts

Log = logging.getLogger("gdb_log")

REGEX_MESSAGE = re.compile(r"^:\w+!\w+@\w+\.tmi\.twitch\.tv PRIVMSG #\w+ :")

REGEX_GREETING = r"(hi|hiya|hello|hey|yo|sup|howdy|hovvdy|greetings|what's good|whats good|vvhat's good|vvhats good|what's up|whats up|vvhat's up|vvhats up|konichiwa|hewwo|etalWave|vvhats crackalackin|whats crackalackin|henlo|good morning|good evening|good afternoon) (@*GamerDeathBot|gdb)"

REGEX_FAREWELL = r"(bye|goodnight|good night|goodbye|good bye|see you|see ya|so long|farewell|later|seeya|ciao|au revoir|bon voyage|peace|in a while crocodile|see you later alligator|later alligator|have a good one|igottago|l8r|later skater|catch you on the flip side|bye-bye|sayonara) (@*GamerDeathBot|gdb)"

def parse_msg(conn, msg, active_channels):
    """Thread to read from the socket connection.

    Args:
        conn -- SocketConnection object
        msg -- message to parse
        active_channels -- active channel transmit dictionary
    """
    if msg is None:
        return

    # Handle ping pong, once ~5 mins
    if msg == "PING :tmi.twitch.tv":
        conn.send("PONG :tmi.twitch.tv")
        return

    # Check for more than one message per recv call
    message_chunks = msg.split("\r\n")
    for split_msg in message_chunks:
        username, message, channel = split_msg_data(split_msg)

        Log.info(str(channel) + " -- " + str(username) + ": " + message.strip())

        # Match a greeting message
        if re.search(REGEX_GREETING, message, re.IGNORECASE):
            active_channels[channel].send_greeting(username)

        # Match a farewell message
        elif re.search(REGEX_FAREWELL, message, re.IGNORECASE):
            active_channels[channel].send_farewell(username)

        # Match a gamerdeath message
        elif message == "!gamerdeath":
            active_channels[channel].send_gamerdeath()
    
def split_msg_data(msg):
    """Split out user, message and channel names from a receieved message

    Args:
        msg -- message to parse
    Return:
        username, message, channel 3-tuple. Strings, but channel can be None.
    """
    username = re.search(r"\w+", msg)
    message = REGEX_MESSAGE.sub("", msg)
    channel = re.search(r"#\w+", msg)

    if username:
        username = username.group(0)

    if channel:
        channel = channel.group(0)[1:]

    return username, message, channel

def handle_sigint(sig, frame):
    """SIGINT signal handler"""
    sys.exit(0)

def setup_logger():
    """Initialize logger to write to /var/log/gdb"""
    global Log
    Log.setLevel(logging.INFO)
    formatter = logging.Formatter('%(asctime)s : %(message)s')
    fileHandler = logging.FileHandler("/var/log/gdb", mode='w')
    fileHandler.setFormatter(formatter)
    streamHandler = logging.StreamHandler()
    streamHandler.setFormatter(formatter)

    Log.addHandler(fileHandler)
    Log.addHandler(streamHandler)

def main():
    signal.signal(signal.SIGINT, handle_sigint)
    setup_logger()

    # Setup the socket connection and the rx thread.
    api = TwitchAPIClient(consts.CLIENT_ID, consts.PASS)
    conn = SocketConnection()

    active_channels = {}
    for chan in consts.TARGET_CHANNELS:
        chan_id = api.get_user_id(chan)
        active_channels[chan] = ChannelTransmit(conn, api, chan, chan_id)

    conn.connect()

    # Rx and process messages forever
    while True:
        parse_msg(conn, conn.recv(), active_channels)

if __name__ == '__main__':
    main()
