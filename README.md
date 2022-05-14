# Rust query proxy

A very simple proxy server for querying data about Rust game servers (Source Engine Query).

## Building from source

1. [Install the Go compiler](https://golang.org/dl/)
2. In the project directory, run `go build`

## Config

Parameters:

| Parameter                    | Default value      | Description                                               |
|------------------------------|--------------------|-----------------------------------------------------------|
| Bind                         | `0.0.0.0:5050`     | Proxy server listen address:port                          |
| IpWhitelist                  | `["127.0.0.1"]`    | Whitelist for incoming connections (array of addresses)   |
| QueryIntervalInSeconds       | `30`               | Time until next game-server update (in ns)                |
| ServerCacheTimeInSeconds     | `60`               | Game-server cache time (in ns)                            |
| QueryConnectTimeoutInSeconds | `5`                | Game-server connection timeout when updating data (in ns) |
| UpdateBurstLimit             | `5`                | Number of simultaneous updates                            |
| CustomTagsWhitelist          | `["monthly", ...]` | Game-server custom tags whitelist                         |

## Connecting

### Client to server payload

`0xFF + ver(2) + mode(1) + count(2) + (ip(4) + port(2))...`

Byte order: Big endian

| Value   | Size              | Notes                                                               |
|---------|-------------------|---------------------------------------------------------------------|
| 0xFF    | 1 byte            | always 0xFF                                                         |
| ver     | 2 bytes           | Version of client (Current server version: 2)                       |
| mode    | 1 byte            | Working mode (Currently only supported values: 0)                   |
| count   | 2 bytes           | Number of game server addresses (0 - 65535)                         |
| ip+port | (4+2 bytes)*count | Game server encoded addresses (4 bytes of IPv4 and 2 bytes of port) |

### Server to client payload

`hostname(str) + 0x00 + map(str) + 0x00 + players(2) + max_players(2) + wiped(8) + queue(2) + tags(string)` (sequence)

| Value       | Size                   | Notes                         |
|-------------|------------------------|-------------------------------|
| hostname    | Zero-terminated string | Hostname of the server        |
| map         | Zero-terminated string | Map of the server             |
| players     | 2 bytes (uint16)       | Current player count          |
| max_players | 2 bytes (uint16)       | Max player count              |
| wiped       | 8 bytes (uint64)       | Server wipe date in unix-time |
| queue       | 2 bytes (uint16)       | Queue player count            |
| tags        | Zero-terminated string | Custom tags                   |
