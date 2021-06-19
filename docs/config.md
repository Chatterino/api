# Configuration

Chatterino API's configuration is manageable via config files, environment variables and command line flags.


## Config files

Chatterino API will attempt reading config file from the following paths:
 - `/etc/chatterino-api/config.yaml`
 - `$XDG_CONFIG_HOME` (default: `~/.config/chatterino-api/config.yaml` or `%APPDATA%\chatterino-api\config.yaml` on Windows)
 - current working directory (`./config.yaml`)

The values from each file is merged. Each subsequent read overrides previously set values (i.e. current working directory config file has priority over `$XDG_CONFIG_HOME` config file.

The [default config file](https://github.com/Chatterino/api/blob/master/config.yaml) has default values. Copy it, uncomment and change values as needed.


## Environment variables

All respected environment variables follow the format `CHATTERINO_API_X`, where X is the uppercased corresponding setting from config file, e.g. `CHATTERINO_API_BASE_URL`.


## Command line flags

All command line flags are the same as settings in config file. You can also run API with `--help` flag for list of those and more detailed information about each of those.
