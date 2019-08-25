#!/usr/bin/python

import socket
import time
from threading import Lock

import env

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
        else:
            return func(*args, **kwargs)
    return wrap

class SocketConnection():
    """Manages the socket connection"""
    def __init__(self):
        """Constructor."""
        self._RX_BUF_SZ = 1024
        self._sock = None
        self.is_connected = False

    def connect(self):
        """Initializes connection to the Twitch servers."""
        if not self.is_connected:
            self._sock = socket.socket()
            self._sock.connect((env.HOST, env.PORT))
            self.is_connected = True

            self.send("PASS " + env.PASS)
            self.send("NICK " + env.USER)
            for chan in env.TARGET_CHANNELS:
                self.send("JOIN " + chan)

    @_check_conn
    def send(self, msg):
        """Send a message on the socket. Will attempt reconnection if socket conn is lost.
    
        Args:
            msg -- string to send
        """
        try:
            self._sock.sendall((msg + "\r\n").encode("utf-8"))
        except Exception as e:
            print(e)
            self.is_connected = False
            time.sleep(1)
            self.connect()

    @_check_conn
    def recv(self):
        """Read data from the socket.

        Returns:
            stripped decoded string
        """
        return self._sock.recv(self._RX_BUF_SZ).decode("utf-8").strip()

    def chat(self, channel, msg):
        """Send a chat message on a channel
    
        Args:
            channel -- channel to send on
            msg -- string to send
        """
        m = "PRIVMSG " + channel + " :" + msg
        print("TX: " + m)
        self.send(m)
