In the root directory there's a `docker-compose.yml` file which defines the Chatterino API, a PostgreSQL database, and Prometheus

You can configure those with the `env` and `chatterino2-api.env` files in this directory.

If you're going to make changes to the env files directly in the repo, you will want to ignore them locally:  
`git update-index --assume-unchanged env chatterino2-api.env`
