import time

import requests
from requests.compat import urljoin

class TwitchAPIClient():
    """Twitch API client."""

    def __init__(self, client_id, oauth_token):
        """Initialize the API."""
        self._client_id = client_id
        self._oauth_token = oauth_token

    def channel_is_live(self, channel_id):
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
        """Perform a HTTP GET request."""
        url = urljoin('https://api.twitch.tv/kraken/', path)
        headers = self._get_request_headers()

        response = requests.get(url, params=params, headers=headers)
        
        if response.status_code >= 500:
            print("Got status " + str(response.status_code))

        response.raise_for_status()
        
        return response.json()