#!/usr/bin/python

import threading
import time

class CommandCooldown:
    def __init__(self, timeout):
        self._lock_event = threading.Event()
        self._timer_thread = threading.Thread(target=self._timer)
        self._timer_thread.daemon = True
        self._timeout = timeout
        self._timer_thread.start()

    def _timer(self):
        while True:
            self._lock_event.wait()
            time.sleep(self._timeout)
            self._lock_event.clear()

    def set_cooldown(self):
        self._lock_event.set()

    def check_cooldown(self):
        return not self._lock_event.is_set()
