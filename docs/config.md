# Configuration

Chatterino API's configuration is manageable via config files, environment variables and command line flags.


## Config files

Chatterino API will attempt reading config file in the following paths:
 - `/etc/chatterino-api/`
 - `$XDG_CONFIG_HOME` (default: `~/.config/chatterino-api/` or `%APPDATA\chatterino-api\` on Windows)
 - current working directory

The [default config file](https://github.com/Chatterino/api/blob/master/config.yaml) has default values. Copy it, uncoumment and change values as needed.


## Environment variables

All respected environment variables follow the format `CHATTERINO_API_X`, where X is the uppercased corresponding setting from config file, e.g. `CHATTERINO_API_BASE_URL`.


## Command line flags

All command line flags are the same as settings in config file. You can also run API with `--help` flag for list of those and more detailed information about each of those.
