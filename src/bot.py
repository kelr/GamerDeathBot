#!/usr/bin/python
"""Main Module"""

import time
import re
import threading
import random
from datetime import datetime

from cooldown import CommandCooldown
from conn import SocketConnection
from api.client import TwitchAPIClient
import consts

# Timer threads
CD_TIMERS = {
    "greeting" : CommandCooldown(10),
    "gamerdeath" : CommandCooldown(60)
}

# Available greetings
GREETS = (
    "Hi",
    "Hello",
    "Hey",
    "Yo",
    "What's up",
    "How's it going",
    "Greetings",
    "vvhat's up",
    "Henlo",
    "Howdy",
    "Hovvdy"
)

MSG_REGEX = re.compile(r"^:\w+!\w+@\w+\.tmi\.twitch\.tv PRIVMSG #\w+ :")
MSG_GREETING = r"(hi|hello|hey|yo|sup|howdy|hovvdy|greetings|what's good|whats good|vvhat's good|vvhats good|what's up|whats up|vvhat's up|vvhats up|konichiwa|hewwo|etalWave|vvhats crackalackin|whats crackalackin|henlo|good morning|good evening|good afternoon) @*GamerDeathBot"

def rx_thread(conn):
    """Thread to read from the socket connection.

    Args:
        conn -- SocketConnection object
    """
    while True:
        response = conn.recv()
        if response is not None:
            parse_msg(conn, response)

def parse_msg(conn, msg):
    """Thread to read from the socket connection.

    Args:
        conn -- SocketConnection object
    """
    # Handle ping pong, once ~5 mins
    if msg == "PING :tmi.twitch.tv":
        conn.send("PONG :tmi.twitch.tv")
        return

    username = re.search(r"\w+", msg).group(0)
    message = MSG_REGEX.sub("", msg)
    channel = re.search(r"#\w+", msg)
    if channel:
        channel = channel.group(0)

    print(str(datetime.now()) + " " + str(channel) + " -- " + username + ": " + message.strip())

    # Match a greeting message
    if re.match(MSG_GREETING, message, re.IGNORECASE):
        send_greeting(conn, channel, username)

    # Match a gamerdeath message
    elif message == "!gamerdeath":
        send_gamerdeath(conn, channel)

def send_greeting(conn, channel, username):
    """Send a greeting message when someone says Hi to GDB.

    Args:
        conn -- SocketConnection object
        channel -- channel to reply to
        username -- username to reply to
    """
    if CD_TIMERS["greeting"].check_cooldown():
        conn.chat(channel, get_random_greeting(username))
        CD_TIMERS["greeting"].set_cooldown()

def send_gamerdeath(conn, channel):
    """Send a gamerdeath message when someone invokes !gamerdeath.

    Args:
        conn -- SocketConnection object
        channel -- channel to reply to
    """
    if CD_TIMERS["gamerdeath"].check_cooldown():
        conn.chat(channel, "MrDestructoid Chat, remember to get up and stretch to prevent Gamer Death!")
        CD_TIMERS["gamerdeath"].set_cooldown()

def get_random_greeting(username):
    """Build a random greeting message.

    Args:
        username -- username to reply to
    """
    return random.sample(GREETS, 1)[0] + " " + username + " etalWave"

def getup_thread(conn, api, channel):
    """Thread to tell the gamers to get up every so often. Check for live every 5min.

    Args:
        conn -- SocketConnection object
        api -- API comm object
        channel -- Channel to monitor
    """
    success_count = 0
    while True:
        if api.channel_is_live(consts.CHANNEL_ID[channel]):
            # Send alert in 3 hours
            if success_count >= 36:
                conn.chat(channel, "MrDestructoid " + channel[1:] + " alert! It's been 3 hours and its time to prevent Gamer Death!")
                success_count = 0
            success_count += 1
            print(success_count)
        else:
            success_count = 0
        time.sleep(300)

def main():
    """Setup the socket connection and the rx thread."""
    api = TwitchAPIClient(consts.CLIENT_ID, consts.PASS)
    conn = SocketConnection()
    conn.connect()

    rx_t = threading.Thread(target=rx_thread, args=(conn,))
    rx_t.daemon = True
    rx_t.start()

    getup_thread_list = []
    for chan in consts.TARGET_CHANNELS:
        tmp = threading.Thread(target=getup_thread, args=(conn, api, chan))
        tmp.daemon = True
        tmp.start()
        getup_thread_list.append(tmp)

    # Do nothing forever
    while True:
        time.sleep(1)

if __name__ == '__main__':
    main()
