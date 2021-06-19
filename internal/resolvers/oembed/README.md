# oEmbed

The oEmbed resolver requires a `providers.json` to be loadable from the application.

The `oembed-providers-path` config option can be set to change where the file is loaded from (defaults to `./data/oembed/providers.json`).

If Facebook and Instagram oEmbed providers are in your `providers.json` file, you can specify the `oembed-facebook-app-id` and `oembed-facebook-app-secret` config options to grant Authorization for those requests, giving you rich data for Facebook and Instagram posts.
