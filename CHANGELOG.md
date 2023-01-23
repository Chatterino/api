# Changelog

## Unreleased

- Breaking: Go version 1.17 is now the minimum required version to build this. (#292)
- Breaking: `enable-lilliput` config renamed to `enable-animated-thumbnails`. (#312)
- Breaking: Thumbnail generation now requires libvips. See [docs/build.md](./docs/build.md) for prerequisite instructions. (#366, #369, #312)
- Breaking: Resolver caches are now stored in PostgreSQL. See [docs/build.md](./docs/build.md) for prerequisite instructions. (#271)
- PDF: Generate customized tooltips for PDF files. (#374, #377)
- Twitter: Generate thumbnails with all images of a tweet. (#373)
- YouTube: Added support for 'YouTube shorts' URLs. (#299)
- Minor: Add ability to opt out hostnames from the API. (#405)
- Fix: SevenTV emotes now resolve correctly. (#281, #288, #307)
- Fix: YouTube videos are no longer resolved as channels. (#284)
- Fix: Default resolver no longer crashes when provided url is broken. (#310)
- Fix: JSON responses now always return the proper content type. (#334)
- Dev: Improve BetterTTV emote tests. (#282)
- Minor: BetterTTV cache key changed from plural to singular form. (#282)
- Dev: Add docker-compose support. (#395)
- Dev: Improve Twitch.tv clip tests. (#283)
- Dev: Improve YouTube tests. (#284)
- Dev: Resolver Check now returns a context. (#287)
- Dev: Improve Wikipedia tests. (#286)
- Dev: Improve Imgur tests. (#289)
- Dev: Improve migration tests. (#290)
- Dev: Improve Twitter tests. (#293)
- Dev: Improve SevenTV tests. (#294)
- Dev: Improve FrankerFaceZ tests. (#295)
- Dev: Improve Livestreamfails tests. (#297, #301)
- Dev: Improve default resolver tests. (#300)
- Dev: Resolve imgur.io links. (#365)
- Dev: Don't use `stampede` for link resolver links. (#394)
- Dev: Update to Twitter's v2 API. (#414)
- Dev: Add HTTP Caching headers. (#417)

## 1.2.3

- Dev: Automatically publish docker image to the GitHub Container Registry. (#279)

## 1.2.2

- YouTube: Added comment count to rich video tooltips. (#252)
- YouTube: Added a red `AGE RESTRICTED` label to the YouTube video tooltip. (#251)
- YouTube: Removed dislike count from rich tooltips since YouTube removed it. (#243)
- Twitter: Blacklist special pages from being resolved as user pages. (#220)
- Twitch: Handle Twitch clips from `m.twitch.tv` domain. (#239)
- Updated Facebook & Instagram endpoints to oembed v10. (#201)
- Added a Chatterino API Privacy Policy and Terms of Service to `/legal/privacy-policy` and `/legal/terms-of-service`. (#253)
- Dev: Disable CodeGQL. (#275)
- Dev: Add CodeCov support. (#276)
- Dev: Add CodeCov badge to readme. (#277)

## 1.2.1

- MaxThumbnailSize is now configurable using the `max-thumbnail-size` config value. (#195)
- Twitch clips under `www.twitch.tv` domain work again. (#189)
- Imgur thumbnails are now proxied as well. (#187)
- Added link preview support for 7tv emote links. (#155)
- Skip lilliput if image is below maxThumbnailSize. (#184)
- Dev: Change Emote Set backend from `twitchemotes.com` to the Twitch Helix API. (#175, #188)

## 1.2.0

- Breaking: YouTube environment variable has been renamed (`CHATTERINO_API_YOUTUBE_API_KEY`).
- Added viper as a configuration manager. This allows to set configuration with config files, environment variables or command line flags. There are also new configurable properties. See docs/config.md for detailed information. (#162)
- Pass http.Request all the way down the pipeline to custom resolvers. (#167)

## 1.1.0

- Made Reddit Score field in Livestreamfails tooltip use humanized value. (#164)
- Added support for customizable oEmbed resolving for websites with the `providers.json` file. See [`data/oembed/providers.json`](data/oembed/providers.json). Three new environment variables can be set. See [`internal/resolvers/oembed/README.md`](internal/resolvers/oembed/README.md) (#139, #152)
- Added support for YouTube channel links. (#157)
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
