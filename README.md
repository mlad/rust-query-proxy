# Rust query proxy

A very simple proxy server for querying data about Rust game servers (Source Engine Query).

## Building from source

1. [Install the Go compiler](https://golang.org/dl/)
2. In the project directory, run `go build`

## Config

Parameters:

| Parameter | Default value | Description
----------- | ------------- | -----------
| address | `0.0.0.0:5050` | Proxy server listen address:port
| whitelist | `[]` | Whitelist for incoming connections (array of addresses)
| server_update_time | `30000000000` (30s) | Time until next game-server update (in ns)
| server_cache_time | `60000000000` (60s) | Game-server cache time (in ns)
| query_connect_timeout | `5000000000` (5s) | Game-server connection timeout when updating data (in ns)
| update_burst_limit | `5` | Number of simultaneous updates

## Connecting

### Client to server payload

`0xFF + ver(2) + mode(1) + count(2) + (ip(4) + port(2))...`

Byte order: Big endian

| Value | Size | Notes |
| ----- | ---- | ----- |
| 0xFF | 1 byte | always 0xFF |
| ver | 2 bytes | Version of client (Current server version: 1) |
| mode | 1 byte | Working mode (Currently only supported values: 0) |
| count | 2 bytes | Number of game server addresses (0 - 65535) |
| ip+port | (4+2 bytes)*count | Game server encoded addresses (4 bytes of IPv4 and 2 bytes of port) |

### Server to client payload

`hostname(str) + 0x00 + map(str) + 0x00 + players(2) + max_players(2) + wiped(8)` (sequence)

| Value | Size | Notes |
| ----- | ---- | ----- |
| hostname | Zero-terminated string | Hostname of the server |
| map | Zero-terminated string | Map of the server |
| players | 2 bytes | Current player count |
| max_players | 2 bytes | Max player count |
| wiped | 8 bytes | Server wipe date in 64 bit unixtime |
