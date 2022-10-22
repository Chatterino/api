[![Build status](https://github.com/Chatterino/api/actions/workflows/build.yml/badge.svg?branch=master)](https://github.com/Chatterino/api/actions/workflows/build.yml?query=branch%3Amaster)
[![codecov](https://codecov.io/gh/Chatterino/api/branch/master/graph/badge.svg?token=gz6EYE3bQQ)](https://codecov.io/gh/Chatterino/api)

# API

Go web service that serves as a cache to APIs that each Chatterino client could use.

## Routes

### Resolve Twitch emote set

`twitchemotes/set/:setID`
Returns information about a given Twitch emote set. Example response:

```javascript
{
  "channel_name": "forsen", // twitch user name
  "channel_id": "22484632", // twitch user id
  "type": "sub",            // string describing what type of emote set this is (e.g. "sub")
  "custom": false           // indicates whether this is added/modified by us or straight passthrough from the twitchemotes API
  "tier": 1,                // integer describing what tier the sub emote is part of
}
```

### Resolve URL

`link_resolver/:url`  
Resolves a url into a preview tooltip.  
Route content type: `application/json`  
Route HTTP Status Code is almost always `200` as long as we were able to generate information about the URL, even if the API we call returns 404 or 500.  
If the given URL is not a valid url, the Route HTTP status code will be `400`.

#### Examples

`url` parameter: `https://example.com/page`

```javascript
{
  "status": 200,                                               // status code returned or inferred from the page
  "thumbnail": "http://api.url/thumbnail/web.com%2Fimage.png", // proxied thumbnail url if there's an image
  "message": "",                                               // used to forward errors in case the website e.g. couldn't load
  "tooltip": "<div>tooltip</div>",                             // HTML tooltip used in Chatterino
  "link": "http://example.com/longer-page"                     // final url, after any redirects
}
```

`url` parameter: `https://example.com/error`

```json
{
  "status": 404,
  "message": "Page not found"
}
```

### API Uptime

`health/uptime`  
Returns API service's uptime. Example response:

```
928h2m53.795354922s
```

### API Memory usage

`health/memory`  
Returns information about memory usage. Example response:

```
Alloc=505 MiB, TotalAlloc=17418866 MiB, Sys=3070 MiB, NumGC=111245
```

### API Uptime and memory usage

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
