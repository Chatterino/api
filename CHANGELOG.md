# Changelog

## Unreleased

- Added support for customizable oEmbed resolving for websites with the `providers.json` file. See [`data/oembed/providers.json`](data/oembed/providers.json). Three new environment variables can be set. See [`internal/resolvers/oembed/README.md`](internal/resolvers/oembed/README.md) (#139, #152)
- Breaking: Environment variable `CHATTERINO_API_CACHE_TWITCH_CLIENT_ID` was renamed to `CHATTERINO_API_TWITCH_CLIENT_ID`. (#144)
- Dev, Breaking: Replaced `dankeroni/gotwitch` with `nicklaw5/helix`. This change requires you to add new environment variable: `CHATTERINO_API_TWITCH_CLIENT_SECRET` - it's a client secret generated for your Twitch application.

## 1.0.2

- Twitter profile pictures are now returned in their original quality. (#131)
- Youtube thumbnails are now in medium quality instead of standard definition to get a rectangular shaped image. (#127)

## 1.0.1

- Dev: You can now set the Base URL flag with the `CHATTERINO_API_BASE_URL` environment variable. (#123)
  Note that the priority will always be CLI Argument > Environment Variable, so passing `-b` to the application will negate any environment variables set.
- Dev: Add Dockerfile (#125)
- Dev: You can now set the Base URL flag with the `CHATTERINO_API_BIND_ADDRESS` environment variable. (#124)
  Note that the priority will always be CLI Argument > Environment Variable.

## 1.0.0

- Non-Windows builds now use `discord/lilliput` to support animated GIFs and static WebP thumbnails. (#119)
- Twitter resolver timestamp date format changed from `Jan 2 2006` to `02 Jan 2006` (#105)
- Dev, Breaking: Moved main package from root directory to `/cmd/api`. This change also changes the path of the executable from `./api` to `./cmd/api/api` - make sure to reflect this change in your systemd unit. (#104, #107)
- Dev: Replaced `gorilla/mux` with `go-chi/chi`. This change requires the URL parameters to be percent-encoded, specifically the slashes. (#99)
- Added support for Livestreamfails clip links. (#98, #112)
- Added support for Wikipedia article links. (#92, #118)
- Fixed an issue where OpenGraph descriptions were not HTML-sanitized. (#90)
- Added support for imgur image links. (#81)
- Added support for FrankerFaceZ emote links. (#57, #110)
- Added author name to Twitch clips response. (#76)
