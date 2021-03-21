# Unreleased

- Dev: Moved main package from root directory to `/cmd/api` (#104)
- Dev: Replaced `gorilla/mux` with `go-chi/chi`. This change requires the URL parameters to be percent-encoded, specifically the slashes. (#99)
- Added support for Livestreamfails clip links. (#98)
- Added support for Wikipedia article links. (#92)
- Fixed an issue where OpenGraph descriptions were not HTML-sanitized. (#90)
- Added support for imgur image links. (#81)
- Added support for FrankerFaceZ emote links. (#57)
- Added author name to Twitch clips response. (#76)
