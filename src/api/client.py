#!/usr/bin/python
"""API Client"""

import requests
from requests.compat import urljoin
import datetime

Log = logging.getLogger("gdb_log")

class TwitchAPIClient():
    """Twitch API client."""

    def __init__(self, client_id, oauth_token):
        """Initialize the API.

        Args:
            client_id -- client id
            oauth_token -- super secret oauth token
        """
        self._client_id = client_id
        self._oauth_token = oauth_token

    def channel_is_live(self, channel_id):
        """Check if a channel is live by ID.

        Args:
            channel_id -- integer ID to check

        Return:
            Boolean.
        """
        params = {
            'stream_type': 'live',
        }
        response = self._request_get('streams/{}'.format(channel_id), params=params)

        return True if response['stream'] else False

    def channel_uptime(self, channel_id):
        """Get a channel's uptime in seconds

        :param channel_id: integer ID to check
        :return:  integer
        """
        try:
            started = self._request_get_new("streams?user_id=%s" % channel_id)["data"][0]["started_at"]
        except IndexError:  # channel is offline or channel doesnt exist
            uptime = -1
        else:
            starttime = datetime.datetime.strptime(started, '%Y-%m-%dT%H:%M:%SZ')
            uptime = (datetime.datetime.utcnow() - starttime).seconds
        return uptime

    def get_user_id(self, username):
        """Get a user id from a username.

        Args:
            username -- username to check

        Return:
            ID string or None
        """
        response = self._request_get('users?login={}'.format(username))

        ident = None
        try:
            ident = response['users'][0]["_id"]
        except IndexError:
            Log.error("Invalid username: " + username)

        return ident

    def _get_request_headers(self):
        """Prepare the headers for the requests."""
        headers = {
            'Accept': 'application/vnd.twitchtv.v5+json',
            'Client-ID': self._client_id,
            'Authorization': "OAuth " + self._oauth_token
        }
        return headers

    def _request_get(self, path, params=None):
        """Perform a HTTP GET request.

        Args:
            path -- Path to append on default api path.
            params -- Extra parameters to append on
        Return:
            JSON object response
        """
        url = urljoin('https://api.twitch.tv/kraken/', path)
        headers = self._get_request_headers()

        response = requests.get(url, params=params, headers=headers)

        if response.status_code >= 500:
            Log.error("Got status " + str(response.status_code))

        response.raise_for_status()

        return response.json()

    def _request_get_new(self, path, params=None):
        """Perform a HTTP GET request.

        Args:
            path -- Path to append on default api path.
            params -- Extra parameters to append on
        Return:
            JSON object response
        """
        url = urljoin('https://api.twitch.tv/helix/', path)
        headers = {
            'Client-ID': self._client_id,
            'Authorization': "Bearer " + self._oauth_token
        }

        response = requests.get(url, params=params, headers=headers)

        if response.status_code >= 500:
            Log.error("Got status " + str(response.status_code))

        response.raise_for_status()

        return response.json()
