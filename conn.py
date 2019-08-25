#!/usr/bin/python

import socket
import time

import env

def check_conn(func):
    def wrap(obj, *args, **kwargs):
        if obj.is_connected:
            return func(obj, *args, **kwargs)
        else:
            print("not conn")
    return wrap

class SocketConnection():
    def __init__(self):
        self._RX_BUF_SZ = 1024
        self._sock = None
        self.is_connected = False

    def connect(self):
        if not self.is_connected:
            self._sock = socket.socket()
            self._sock.connect((env.HOST, env.PORT))
            self.is_connected = True

            self.send("PASS " + env.PASS)
            self.send("NICK " + env.USER)
            for chan in env.TARGET_CHANNELS:
                self.send("JOIN " + chan)

    @check_conn
    def send(self, msg):
        try:
            self._sock.send((msg + "\r\n").encode("utf-8"))
        except Exception as e:
            print(e)
            self.is_connected = False
            time.sleep(1)
            self.connect()

    @check_conn
    def recv(self):
        m = self._sock.recv(self._RX_BUF_SZ).decode("utf-8").strip()
        return m

    def chat(self, channel, msg):
        m = "PRIVMSG " + channel + " :" + msg
        print("TX: " + m)
        self.send(m)
