[![CircleCI](https://circleci.com/gh/Chatterino/api.svg?style=svg)](https://circleci.com/gh/Chatterino/api)

# chatterino-api-cache

simple go web service that serves as a cache to APIs that chatterino wants to use.

Emote data is served cached from [twitchemotes.com](https://twitchemotes.com/).

## routes
`/twitchemotes/set/:setID`  
returns information about a given twitch emote set. Response example:
```
{
    "channel_name": "forsen", // twitch user name
    "channel_id": "22484632", // twitch user id
    "type": "sub",            // string describing what type of emote set this is (e.g. "sub")
    "custom": false           // indicates whether this is added/modified by us or straight passthrough from the twitchemotes api
}
```
