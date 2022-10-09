# How to build and host the Chatterino 2 API

## Prerequisites

1. Resolved links are stored in PostgreSQL, so you must have PostgreSQL installed and accessible for the user running the API. For Ubuntu, you would install it with `sudo apt install postgresql`, create a DB user for your system user (`sudo -upostgres createuser pajlada`), then create a db for the api (`sudo -upostgres createdb chatterino-api --owner pajlada`). Make sure to edit `dsn` in your [configuration](./config.md). Example, using the details above, `dsn:"host=/var/run/postgresql user=pajlada database=chatterino-api"`.
2. You must have [`libvips`](https://github.com/libvips/libvips) >=8.12.0 installed for thumbnail generation.

   On Ubuntu 22.04, this can be done with `sudo apt install libvips libvips-dev`.
   Different distros or releases may require adding a PPA or building and installing from source.

   For Windows when getting the [Windows binaries](https://github.com/libvips/build-win64-mxe/releases/latest) from the libvips link above make sure to get the ones suffixed with "-static".
   Install pkg-config via choco `choco upgrade -y pkgconfiglite` and setup the following environment variables:
   `VIPS_PATH` to the directory where vips binaries are downloaded and extracted,
   `PKG_CONFIG_PATH` to `%VIPS_PATH%\lib\pkgconfig`
   and append `%VIPS_PATH%\bin` to your `Path` environment variable.

## Build

1. Clone the repo: `git pull https://github.com/Chatterino/api.git`
1. Move into the directory, fetch the dependencies and build the binary: `cd api && make build`
1. Edit API's configuration, see [configuration](./config.md)
1. Start the API: `./cmd/api/api`

## Install systemd unit

1. Install the pre-packaged systemd unit file: `sudo cp ./docs/chatterino-api.service /etc/systemd/system/`
1. Use your preferred editor to edit the service file with all the necessary details (incl. [API keys](./apikeys.md)): `sudo editor /etc/systemd/system/chatterino-api.service`
1. Tell systemd to reload the changes: `sudo systemctl daemon-reload`
1. Start the service and enable to start on boot: `sudo systemctl enable --now chatterino-api`
