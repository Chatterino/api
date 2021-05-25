# oEmbed

The oEmbed resolver requires a `providers.json` to be loadable from the application.

The `CHATTERINO_API_OEMBED_PROVIDERS_PATH` environment variable can be set to change where the file is loaded from, and if no environment variable is set it tries to load the file from `./providers.json`.

If Facebook and Instagram oEmbed providers are in your `providers.json` file, you can specify the `CHATTERINO_API_OEMBED_FACEBOOK_APP_ID` and `CHATTERINO_API_OEMBED_FACEBOOK_APP_SECRET` environment variables to grant Authorization for those requests, giving you rich data for Facebook and Instagram posts.
