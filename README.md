# usg-dns-api

`usg-dns-api` is a program that exposes an API for managing DNS records on Ubiquiti routers. It interacts with the Unifi Controller and builds a _hosts_ file based on the following:

- Static IP addresses configured by the router administrator.
- DNS records defined via the API.

Reverse records are also automatically generated.

## Installation

1. Download the `.deb` file from the GitHub releases using `curl`:

   ```shell
   curl -L https://github.com/rclsilver-org/usg-dns-api/releases/download/<version>/usg-dns-api_<version>_mips.deb -o usg-dns-api_mips.deb
   ```

2. Install the `.deb` package:

   ```shell
   sudo dpkg -i usg-dns-api_mips.deb
   ```

3. Edit the configuration file located at `/etc/usg-dns-api/usg-dns-api.yaml` to suit your environment.

4. Generate a token for API access:

   ```shell
   sudo usg-dns-api generate-token
   ```

5. Enable the service to start at boot:

   ```shell
   sudo update-rc.d usg-dns-api defaults
   ```

6. Start the service:
   ```shell
   sudo /etc/init.d/usg-dns-api start
   ```

## Configuration

The configuration file is a list of key/value pairs. Every setting may also be
read from the environment by starting the program with `-c ""` and setting
`CONFIGURATION_FROM=env:CFG`, each key being then prefixed with `CFG_`.

| Key | Required | Default | Description |
| --- | --- | --- | --- |
| `HTTP_LISTEN_HOST` | no | `localhost` | Address the API listens on. |
| `HTTP_LISTEN_PORT` | no | `8080` | Port the API listens on. |
| `UNIFI_URL` | yes | | Base URL of the controller, without a trailing slash. |
| `UNIFI_SITE` | yes | | Site name, usually `default`. |
| `UNIFI_API_KEY` | no | | API key authenticating against UniFi OS. |
| `UNIFI_USERNAME` | only without an API key | | Controller account. |
| `UNIFI_PASSWORD` | only without an API key | | Password of that account. |
| `UNIFI_INSECURE` | no | `false` | Skip the TLS certificate verification. |
| `HOSTS_FILE` | no | `/config/user-data/hosts` | Generated hosts file. |
| `DB_PATH` | no | `/config/user-data/usg-dns-api.db` | Database file. |

The last two defaults are the ones baked into the `.deb` package; a local build
writes `hosts` and `usg-dns-api.db` in the current directory instead.

### Authenticating against the controller

Two modes are supported, and the API prefix is selected accordingly.

**API key** — the recommended mode for a UniFi OS console (UDM, UNVR, Cloud Key
Gen2+, UniFi OS Server). Create the key from the UniFi OS interface, under
`Settings > Integrations`, then configure:

```yaml
- key: UNIFI_URL
  value: https://unifi.example.com
- key: UNIFI_API_KEY
  value: your-api-key
```

Each request then carries an `X-API-KEY` header: no session is opened, and
`UNIFI_USERNAME` and `UNIFI_PASSWORD` are ignored.

**Username and password** — kept for the standalone network application:

```yaml
- key: UNIFI_USERNAME
  value: usg-dns-api
- key: UNIFI_PASSWORD
  value: s3cr3t
```

The login is attempted on `/api/auth/login`, which UniFi OS serves, before
falling back to `/api/login` of a standalone controller.

### Self-signed certificate

Controllers ship with a self-signed certificate, which the HTTP client rejects
with a `x509: certificate is valid for ...` error. Skip the verification with:

```yaml
- key: UNIFI_INSECURE
  value: "true"
```

A warning is logged at startup whenever the verification is disabled.

## Persistent dnsmasq Configuration

To configure `dnsmasq` persistently on Ubiquiti routers, edit the `config.gateway.json` file and add the following configuration:

```json
{
  "service": {
    "forwarding": {
      "options": [
        "server=8.8.8.8",
        "server=8.8.4.4",
        "all-servers",
        "no-hosts",
        "addn-hosts=/config/user-data/hosts",
        "domain-needed",
        "bogus-priv",
        "expand-hosts",
        "domain=local.example.com",
        "local=/local.example.com/"
      ]
    }
  }
}
```

## API Usage Examples

- **List all DNS records**:

  ```shell
  curl -i -H "Authorization: <master-token>" http://<router>:8080/records
  ```

- **Add a DNS record**:

  ```shell
  curl -i -H "Authorization: <master-token>" -X POST http://<router>:8080/records -d '{"name": "foo", "target": "127.0.0.1"}'
  ```

- **Update a DNS record**:

  ```shell
  curl -i -H "Authorization: <master-token>" -X PUT http://<router>:8080/records/<id> -d '{"name": "foo", "target": "127.0.0.1"}'
  ```

- **Delete a DNS record**:
  ```shell
  curl -i -H "Authorization: <master-token>" -X DELETE http://<router>:8080/records/<id>
  ```

This API allows you to easily manage DNS records through a simple HTTP interface with the token-based authentication for secure access.
