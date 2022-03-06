This directory contains custom resolvers

When adding a resolver for a new service, you will need to create a new directory (package) here.

The new directory should be named after the service (e.g. `twitch` or `twitter`).

The new directory should contain an `initialize.go` file that contains a public Initialize function which must be called from `internal/resolvers/default/link_resolver.go New`

Each service must contain at least one Resolver+Loader pair.

The Resolver is responsible for checking if a URL should be checked by this package, and interacting with the Loader.

The Loader is responsible for making the API request and formatting the response to a tooltip in case it doesn't already exist in the cache.

For an example of a well structure package, check out the `betterttv` directory.
