#!/usr/bin/python
"""Channel Transmit Control"""

import time
import threading
import random

from cooldown import CommandCooldown
import consts

class ChannelTransmit():
    """Handles transmissions and timers to connected channels."""

    def __init__(self, conn, api, channel_name, channel_id):
        """Constructor.

        Args:
            conn -- SocketConnection object
            api -- API interface object
            channel_name -- Name of the channel to transmit to
            channdl_id -- Id of the channel to transmit to
        """
        self.channel = channel_name
        self.channel_id = channel_id
        self.conn = conn
        self.api = api

        self._timers = {
            "greeting" : CommandCooldown(10),
            "farewell" : CommandCooldown(10),
            "gamerdeath" : CommandCooldown(60)
        }

        self._tx_thread = threading.Thread(target=self._getup_thread)
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

    def send_farewell(self, username):
        """Send a farewell message when someone says bye to GDB.

        Args:
            username -- username to reply to
        """
        if self._timers["farewell"].check_cooldown():
            self.conn.chat(self.channel, self._get_random_farewell(username))
            self._timers["farewell"].set_cooldown()

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
        response = random.sample(consts.GREETING_RESPONSES, 1)[0] + " " + username + " etalWave"
        if username.lower() == "evanito":
            response = response + " You are my favorite chatter :)"
        if username.lower() in consts.SUB_GIFTERS:
            response = response + " Thank you for the sub btw! :)"
        return response

    def _get_random_farewell(self, username):
        """Build a random goodbye message.

        Args:
            username -- username to reply to
        """
        return random.sample(consts.FAREWELL_RESPONSES, 1)[0] + " " + username + " etalWave"

    def _getup_thread(self):
        """Thread to tell the gamers to get up every so often. Check for live every reminder_period."""
        success_count = 0
        reminder_period = 10800  # time in seconds to remind. default: 10800s = 3hrs
        while True:
            uptime = self.api.channel_uptime(self.channel_id)
            if uptime != -1:
                # Send alert in 3 hours
                if int(uptime / reminder_period) > success_count:
                    self.conn.chat(self.channel, "MrDestructoid " + self.channel + " alert! It's been %s hours and its time to prevent Gamer Death!" % str(int(reminder_period / 3600)))
                    success_count = int(uptime / reminder_period)
                print(success_count)
                wait_time = reminder_period - (uptime % reminder_period)
            else:
                success_count = 0
                wait_time = 300
            if wait_time < 5:
                wait_time = 5
            time.sleep(wait_time)  # sleep until the expected time, then check again
