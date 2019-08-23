[![CircleCI](https://circleci.com/gh/Chatterino/api.svg?style=svg)](https://circleci.com/gh/Chatterino/api)

# chatterino-api-cache

simple go web service that serves as a cache to APIs that chatterino wants to use.

Emote data is served cached from [twitchemotes.com](https://twitchemotes.com/).

## Routes
`/twitchemotes/set/:setID`  
returns information about a given twitch emote set. Example response:
```
{
    "channel_name": "forsen", // twitch user name
    "channel_id": "22484632", // twitch user id
    "type": "sub",            // string describing what type of emote set this is (e.g. "sub")
    "custom": false           // indicates whether this is added/modified by us or straight passthrough from the twitchemotes api
}
```

`link_resolver/:url`  
resolve a url into a preview tooltip. Example response:
```
{
    "status": 200,                     // status code returned from the page
    "message": "",                     // used to forward errors in case the website e.g. couldn't load
    "tooltip": "<div>tooltip</div>",   // HTML tooltip used in Chatterino
    "link": "http://final.url.com/asd" // final url, after any redirects
}
```

`health/uptime`  
to be filled in

`health/memory`  
to be filled in

`health/combined`  
to be filled in

## Using your self-hosted version
If you host your own version of this api, you can modify which url Chatterino2 uses to resolve links and to resolve twitch emote sets.  
[Change link resolver](https://github.com/Chatterino/chatterino2/blob/master/docs/ENV.md#chatterino2_link_resolver_url)  
[Change twitch emote resolver](https://github.com/Chatterino/chatterino2/blob/master/docs/ENV.md#chatterino2_twitch_emote_set_resolver_url)  
