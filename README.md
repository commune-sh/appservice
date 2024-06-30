### Commune Appservice

This is a WIP appservice for making matrix rooms publicly accessible - intended
to be used with the Commune client.

#### Configuration

Register a new appservice on your homeserver:

```yaml
id: "commune_public_access"
url: "http://localhost:8989"
as_token: "app_service_access_token"
hs_token: "homeserver_access_token"
sender_localpart: "commune_public_access" 
rate_limited: false
namespaces:
  rooms:
  - exclusive: false
    regex: "!.*:.*"
```

Copy `config.sample.yaml` to `config.yaml` and fill in the required fields.

```toml
[app]
domain = "public.commune.sh"
port = 8989

[appservice]
id = "commune_public_access"
sender_localpart = "commune_public_access"
access_token = "app_service_access_token"
hs_access_token = "homeserver_access_token"

[matrix]
homeserver = "http://localhost:8008"
server_name = "commune.sh"

[security]
allowed_origins = ["http://public.commune.sh"]

[log]
max_size = 100
max_backups = 7
max_age = 30
compress = true
```

Run `make` to build the binary `./bin/commune`.
