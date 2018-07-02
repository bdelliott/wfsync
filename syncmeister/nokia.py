import os
import time
import uuid


def cli():
    """Entry point for CLI testing of Nokia API"""

    # application (consumer) credentials:
    api_key = os.environ['NOKIA_API_KEY']
    api_secret = os.environ['NOKIA_API_SECRET']

    # https://developer.health.nokia.com/api/doc#api-OAuth_Authentication

    # Step 1 -
    #   "Generate an oAuth token to be used for the End-User authorization call."
    #
    url = 'https://developer.health.nokia.com/account/request_token'

    params = [
        ('oauth_callback', 'http://www.elliottsoft.com/syncmeister/nokia/callback'),
        ('oauth_consumer_key', api_key),
        ('oauth_nonce', str(uuid.uuid4())),
        ('oauth_signature_method', 'HMAC-SHA1'),
        ('oauth_timestamp', int(time.time())),
        ('oauth_version', '1.0')
    ]

    # http://requests-oauthlib.readthedocs.io/en/latest/oauth1_workflow.html
    from requests_oauthlib import OAuth1Session
    oauth = OAuth1Session(api_key, client_secret=api_secret)
    resp = oauth.fetch_request_token(url)
    print(resp)

    import pdb; pdb.set_trace()
