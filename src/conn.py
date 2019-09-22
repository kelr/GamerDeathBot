#!/usr/bin/python
"""IRC Server Socket Connection Class"""

import socket
import time

import consts

def _check_conn(func):
    """Decorator to check that a connection is established before doing anything

    Args:
        func -- Function object to wrap
    Returns:
        Wrapped function
    """
    def wrap(*args, **kwargs):
        if not args[0].is_connected:
            raise Exception
        return func(*args, **kwargs)
    return wrap

class SocketConnection():
    """Manages the socket connection"""
    def __init__(self):
        """Constructor."""
        self._RX_BUF_SZ = 2048
        self._sock = None
        self.is_connected = False

    def connect(self):
        """Initializes connection to the Twitch servers."""
        if not self.is_connected:
            self._sock = socket.socket()
            self._sock.settimeout(0.5)
            self._sock.connect((consts.HOST, consts.PORT))
            self.is_connected = True

            self.send("PASS " + consts.PASS)
            self.send("NICK " + consts.USER)
            for chan in consts.TARGET_CHANNELS:
                print("Joining: " + chan)
                self.send("JOIN #" + chan)

    @_check_conn
    def send(self, msg):
        """Send a message on the socket. Will attempt reconnection if socket conn is lost.

        Args:
            msg -- string to send
        """
        try:
            self._sock.sendall((msg + "\r\n").encode("utf-8"))
        except socket.error as ex:
            print(ex)
            self.is_connected = False
            time.sleep(1)
            self.connect()

    @_check_conn
    def recv(self):
        """Read data from the socket. Timeout every 0.5 seconds.

        Returns:
            stripped decoded string or None if no data was recved
        """
        data = None
        try:
            data = self._sock.recv(self._RX_BUF_SZ).decode("utf-8").strip()
        except socket.timeout:
            pass
        return data

    def chat(self, channel, message):
        """Send a chat message on a channel

        Args:
            channel -- channel to send on
            message -- string to send
        """
        msg = "PRIVMSG #" + channel + " :" + message
        print("TX: " + msg)
        self.send(msg)
