# Running in Docker

In the root directory there's a `docker-compose.yml` file which defines the Chatterino API, a PostgreSQL database, and Prometheus

You can configure those with the `env` and `chatterino2-api.env` files in this directory.

If you're going to make changes to the env files directly in the repo, you will want to ignore them locally:  
`git update-index --assume-unchanged env chatterino2-api.env`

## Forwarding prometheus

If you want to forward prometheus at a different route, you'll need to add a `--web.route-prefix` to the command in the `docker-compose.yml` file. For example:  
```diff
     command:
       - '--config.file=/etc/prometheus/prometheus.yml'
       - '--storage.tsdb.path=/prometheus'
+      - '--web.route-prefix=/prometheus'
```
