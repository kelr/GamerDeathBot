#!/usr/bin/python
"""Cooldown Class"""

import threading
import time

class CommandCooldown:
    """Timer thread to reset an event after a certain amount of time."""
    def __init__(self, timeout):
        """Constructor.

        Args:
            timeout -- Time to wait in seconds
        """
        self._lock_event = threading.Event()
        self._timeout = timeout
        self._timer_thread = threading.Thread(target=self._timer)
        self._timer_thread.daemon = True
        self._timer_thread.start()

    def set_cooldown(self):
        """Initialize the cooldown."""
        self._lock_event.set()

    def check_cooldown(self):
        """Check if the thread is in cooldown.

        Returns:
            Boolean. True if is still in cooldown. False otherwise.
        """
        return not self._lock_event.is_set()

    def _timer(self):
        """Timer thread."""
        while True:
            self._lock_event.wait()
            time.sleep(self._timeout)
            self._lock_event.clear()
