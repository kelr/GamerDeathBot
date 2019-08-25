#!/usr/bin/python

import time
import re
import threading
import random

from cooldown import CommandCooldown
from conn import SocketConnection

# Timer threads
CD_TIMERS = {
    "greeting" : CommandCooldown(5),
    "gamerdeath" : CommandCooldown(30)
}

# Available greetings
GREETS = (
    "Hi", 
    "Hello", 
    "Hey", 
    "Yo", 
    "What's up", 
    "How's it going", 
    "Greetings"
)

MSG_REGEX = re.compile(r"^:\w+!\w+@\w+\.tmi\.twitch\.tv PRIVMSG #\w+ :")
MSG_GREETING = r"^(hi|hello|hey|yo|sup|greetings|) @*GamerDeathBot"

def rx_thread(conn):
    while True:
        response = conn.recv()
        if response is None:
            continue
        elif response == "PING :tmi.twitch.tv":
            # Handle ping pong, once ~5 mins
            conn.send("PONG :tmi.twitch.tv")
        else:
            parse_msg(conn, response)

def parse_msg(conn, msg):
    username = re.search(r"\w+", msg).group(0)
    message = MSG_REGEX.sub("", msg)
    channel = re.search(r"#\w+", msg)
    if channel:
        channel = channel.group(0)

    print("RX: " + str(channel) + ":" + username + ":" + message.strip())

    if (re.match(MSG_GREETING, message, re.IGNORECASE)):
        send_greeting(conn, channel, username)

    elif (message == "!gamerdeath"):
        send_gamerdeath(conn, channel)

def send_greeting(conn, channel, username):
    if CD_TIMERS["greeting"].check_cooldown():
        conn.chat(channel, get_random_greeting(username))
        CD_TIMERS["greeting"].set_cooldown()

def send_gamerdeath(conn, channel):
    if CD_TIMERS["gamerdeath"].check_cooldown():
        conn.chat(channel, "MrDestructoid Manual Gamer Death Prevention System activated! Please get up and prevent Gamer Death!")
        CD_TIMERS["gamerdeath"].set_cooldown()

def get_random_greeting(username):
    return random.sample(GREETS, 1)[0] + " " + username + " etalWave"

def main():
    conn = SocketConnection()
    conn.connect()

    rx_t = threading.Thread(target=rx_thread, args=(conn,))
    rx_t.daemon = True
    rx_t.start()

    while True:
        time.sleep(7200)
        for chan in TARGET_CHANNELS:
            chat(sock, chan, "MrDestructoid " + chan[1:] + " alert! It's been 2 hours and you should get up and stretch to prevent Gamer Death!")

if __name__ == '__main__':
    main()
