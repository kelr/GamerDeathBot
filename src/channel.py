#!/usr/bin/python
"""Channel Transmit Control"""

import time
import threading
import random

from cooldown import CommandCooldown
import consts

class ChannelTransmit():
    """Handles transmissions and timers to connected channels."""

    def __init__(self, conn, api, channel_name):
        """Constructor.

        Args:
            conn -- SocketConnection object
            channel -- channel to reply to
        """
        self.channel = channel_name
        self.conn = conn
        self.api = api

        self._timers = {
            "greeting" : CommandCooldown(10),
            "gamerdeath" : CommandCooldown(60)
        }

        self._tx_thread = threading.Thread(target=self._getup_thread())
        self._tx_thread.daemon = True
        self._tx_thread.start()

    def send_greeting(self, username):
        """Send a greeting message when someone says Hi to GDB.

        Args:
            username -- username to reply to
        """
        if self._timers["greeting"].check_cooldown():
            self.conn.chat(self.channel, self._get_random_greeting(username))
            self._timers["greeting"].set_cooldown()

    def send_gamerdeath(self):
        """Send a gamerdeath message when someone invokes !gamerdeath."""
        if self._timers["gamerdeath"].check_cooldown():
            self.conn.chat(self.channel, "MrDestructoid Chat, remember to get up and stretch to prevent Gamer Death!")
            self._timers["gamerdeath"].set_cooldown()

    def _get_random_greeting(self, username):
        """Build a random greeting message.

        Args:
            username -- username to reply to
        """
        return random.sample(consts.GREETING_RESPONSES, 1)[0] + " " + username + " etalWave"

    def _getup_thread(self):
        """Thread to tell the gamers to get up every so often. Check for live every 5min."""
        success_count = 0
        while True:
            if self.api.channel_is_live(consts.TARGET_CHANNELS[self.channel]):
                # Send alert in 3 hours
                if success_count >= 31:
                    self.conn.chat(self.channel, "MrDestructoid " + self.channel[1:] + " alert! It's been 3 hours and its time to prevent Gamer Death!")
                    success_count = 0
                success_count += 1
                print(success_count)
            else:
                success_count = 0
            time.sleep(300)
