[About](#influxdb2-token-manipulator) \| [Usage](#usage)

## Influxdb2 Token Manipulator

This is a tiny helper utility that open's influxdb2's bolt database and replaces authentication
token with a predefined set of secrets, and updates the index afterwards. This is needed to
allow declarative token provisioning in NixOS.

## Usage

Build the utility simply by running:

```bash
$ go build
```

Tokens are identified via a 32-char hexadecimal identifier that has to be placed
anywhere in the token's description. Before building, add the mapping of tokens to
paths containing their desired secret to a `mappings.json` file:

```json
{
  "deadbeef12345678deadbeef12345678": "/path/to/file/containing/new/secret",
  "cafecafe00000000cafecafe00000000": "/path/to/other/file/containing/new/secret",
  # ...
}
```

After a token is created in influxdb, stop influxdb, run the manipulator and start it again.
In NixOS this will be automated using the systemd service.
