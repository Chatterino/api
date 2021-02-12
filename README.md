[![Build Status](https://img.shields.io/endpoint.svg?url=https%3A%2F%2Factions-badge.atrox.dev%2FChatterino%2Fapi%2Fbadge&style=flat)](https://actions-badge.atrox.dev/Chatterino/api/goto)

# API

Go web service that serves as a cache to APIs that each Chatterino client could use.

Emote data is served cached from [twitchemotes.com](https://twitchemotes.com/).

## Routes
`/twitchemotes/set/:setID/`  
Returns information about a given twitch emote set. Example response:
```
{
    "channel_name": "forsen", // twitch user name
    "channel_id": "22484632", // twitch user id
    "type": "sub",            // string describing what type of emote set this is (e.g. "sub")
    "custom": false           // indicates whether this is added/modified by us or straight passthrough from the twitchemotes API
    "tier": 1,                // integer describing what tier the sub emote is part of
}
```

`link_resolver/:url`  
Resolves a url into a preview tooltip. Example response:
```
{
    "status": 200,                                               // status code returned from the page
    "thumbnail": "http://api.url/thumbnail/web.com%2Fimage.png", // proxied thumbnail url if there's an image
    "message": "",                                               // used to forward errors in case the website e.g. couldn't load
    "tooltip": "<div>tooltip</div>",                             // HTML tooltip used in Chatterino
    "link": "http://final.url.com/asd"                           // final url, after any redirects
}
```

`health/uptime`  
Returns API service's uptime. Example response:
```
928h2m53.795354922s
```

`health/memory`  
Returns information about memory usage. Example response:
```
Alloc=505 MiB, TotalAlloc=17418866 MiB, Sys=3070 MiB, NumGC=111245
```

`health/combined`  
Returns both uptime and information about memory usage. Example response:
```
Uptime: 928h5m7.937821282s - Memory: Alloc=510 MiB, TotalAlloc=17419213 MiB, Sys=3070 MiB, NumGC=111246
```

## Using your self-hosted version
If you host your own version of this API, you can modify which url Chatterino2 uses to resolve links and to resolve twitch emote sets.  
[Change link resolver](https://wiki.chatterino.com/Environment%20Variables/#chatterino2_link_resolver_url)  
[Change Twitch emote resolver](https://wiki.chatterino.com/Environment%20Variables/#chatterino2_twitch_emote_set_resolver_url)  
[How to build and host](docs/build.md)
