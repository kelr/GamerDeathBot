#!/usr/bin/python
"""Configuration constants"""

# Twitch Connection Info
HOST = "irc.twitch.tv"
PORT = 6667

# Login info
USER = "gamerdeathbot"
CLIENT_ID = "qx5lh8m0pmmro5uh84z6n19il268a0"
PASS = ""

# Channels to connect to
TARGET_CHANNELS = (
    "#kyrotobi",
    "#gamerdeathbot",
    "#etalyx"
)

CHANNEL_ID = {
    "#kyrotobi" : 31903323,
    "#gamerdeathbot" : 456787927,
    "#etalyx" : 28054687
}

REGEX_MESSAGE = re.compile(r"^:\w+!\w+@\w+\.tmi\.twitch\.tv PRIVMSG #\w+ :")

REGEX_GREETING = r"(hi|hello|hey|yo|sup|howdy|hovvdy|greetings|what's good|whats good|vvhat's good|vvhats good|what's up|whats up|vvhat's up|vvhats up) @*GamerDeathBot"
