# Changelog

## Unreleased

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
