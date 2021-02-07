# How to build and host the Chatterino 2 API

## Build

1. Clone the repo: `git pull https://github.com/Chatterino/api.git`
2. Move into the directory, fetch the dependencies and build the binary: `cd api && make build`
3. Start the API: `./api`

## Install systemd unit
1. Install the pre-packaged systemd unit file: `sudo cp ./docs/example.service /etc/systemd/system/chatterinoapi.service`
2. Use your preferred editor to edit the service file with all the necessary details (incl. [API keys](https://github.com/Chatterino/api/blob/master/docs/apikeys.md)): `sudo editor /etc/systemd/system/chatterinoapi.service`
3. Tell systemd to reload the changes: `sudo systemctl daemon-reload`
4. Start the service and enable to start on boot: `sudo systemctl enable --now chatterinoapi`