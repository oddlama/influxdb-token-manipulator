[Building](#building) \| [Installation and Usage](#installation-and-usage-on-nixos)

## Influxdb Token Manipulator

This is a tiny helper utility that open's influxdb2's bolt database and replaces authentication
token with a predefined set of secrets, and updates the index afterwards. This is needed to
allow declarative token provisioning in NixOS.

## Usage

Tokens are identified via a 32-char hexadecimal identifier that has to be placed
anywhere in the token's description. Before building, add the mapping of tokens to secret values
after the comment in `main.go`:

```go
var tokenPaths = map[string]string{
  // Add token secrets here or in separate file
  "deadbeef12345678deadbeef12345678": "/path/to/file/containing/new/secret",
  // ... add as many definitions as you need
}
```

Now you can build the utility simply by running:

```bash
$ go build
```

After a token is created in influxdb, stop influxdb, run the manipulator and start it again.
In NixOS this will be automated using the systemd service.
