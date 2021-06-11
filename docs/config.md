# Configuration

Chatterino API's configuration is manageable via config files, environment variables and command line flags.


## Config files

Copy the [default config file](https://github.com/Chatterino/api/blob/master/chatterino-api.yaml.dist) to `chatterino-api.yaml`. Chatterino API will attempt reading config file in the following paths:
 - `/etc/`
 - `$XDG_CONFIG_HOME` (default: `~/.config/chatterino-api/`)
 - current working directory


## Environment variables

All respected environment variables follow the format `CHATTERINO_API_X`, where X is the uppercased corresponding setting from config file, e.g. `CHATTERINO_API_BASE_URL`.


## Command line flags

All command line flags are the same as settings in config file. You can also run api with `--help` flag for list of those and more detailed information about each of those.
