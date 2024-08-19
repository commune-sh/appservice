### Commune Appservice

This is an appservice for making matrix rooms and spaces publicly accessible - intended
to be used with [Commune](https://github.com/commune-sh/commune).

The appservice user joins any public matrix rooms it's invited to, and the server proxies specific read-only endpoints to the homeserver's REST API, using the appservice token. 

#### Discovery

The Commune client queries the matrix homeserver's `/.well-known/matrix/client` endpoint to detect whether this appservice is running. Ensure that the endpoint returns the `commune.appservice` URL:

```json
{
  "m.homeserver": {
    "base_url": "https://matrix.commune.sh"
  },
  "commune.appservice": {
    "url": "https://public.commune.sh"
  },
}
```

#### Configuration

Register a new appservice on your Synapse homeserver:

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

For alternative server implementations like Dendrite or Conduit, look up the relevant appservice configuration documentation.

Copy `config.sample.yaml` to `config.yaml` and fill in the required fields.

```toml
[app]
domain = "localhost:8989"
port = 8989

[appservice]
id = "commune"
sender_localpart = "public"
access_token = "app_service_access_token"
hs_access_token = "homeserver_access_token"

[appservice.rules]
auto_join = true
invite_by_local_user = true
federation_domain_whitelist = ["matrix.org", "dev.commune.sh"]

[matrix]
homeserver = "http://localhost:8080"
server_name = "localhost:8480"

[redis]
address = "localhost:6379"
password = ""
rooms_db = 1
messages_db = 2
events_db = 3
state_db = 4

[cache.public_rooms]
enabled = true
expire_after = 14400

[cache.room_state]
enabled = true
expire_after = 3600

[cache.messages]
enabled = true
expire_after = 3600

[log]
max_size = 100
max_backups = 7
max_age = 30
compress = true

```

To ensure that this appservice only joins local homeserver rooms, leave the `federation_domain_whitelist` value empty. 

#### Running

Run `make` to build the binary `./bin/commune`.

#### Deploying

For simplicity, run this appservice on the same host where the matrix homeserver lives, although it isn't necessary. There are example docs for both a systemd unit and nginx reverse proxy in the [`/docs`](https://github.com/commune-sh/appservice/tree/main/docs).

### Development

To develop this appservice, you'll need to have a matrix homeserver running locally. Run `modd` to watch for changes and rebuild the binary.


#### Community

To keep up to date with Commune development, you can find us on `#commune:commune.sh` or `#commune:matrix.org`.

