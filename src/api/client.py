#!/usr/bin/python
"""API Client"""

import requests
from requests.compat import urljoin

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
            print("Got status " + str(response.status_code))

        response.raise_for_status()

        return response.json()
