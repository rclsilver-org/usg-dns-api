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
