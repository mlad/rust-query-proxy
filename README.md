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
| SteamApiToken                | ``                 | Steam API token for loading all servers list              |
| SteamCacheTimeInSeconds      | `300`              | Cache time for server list                                |

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

`ip(4) + port(2) + hostname(str) + 0x00 + map(str) + 0x00 + players(2) + max_players(2) + wiped(8) + queue(2) + tags(string)` (
sequence)

| Value       | Size                   | Notes                               |
|-------------|------------------------|-------------------------------------|
| ip+port     | 4+2 bytes              | Server address (similar to request) |
| hostname    | Zero-terminated string | Hostname of the server              |
| map         | Zero-terminated string | Map of the server                   |
| players     | 2 bytes (uint16)       | Current player count                |
| max_players | 2 bytes (uint16)       | Max player count                    |
| wiped       | 8 bytes (uint64)       | Server wipe date in unix-time       |
| queue       | 2 bytes (uint16)       | Queue player count                  |
| tags        | Zero-terminated string | Custom tags                         |

### Client example

```go
package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

const ProxyAddress = "127.0.0.1:5050"
const ServersCount = 2

func main() {
	conn, err := net.Dial("tcp", ProxyAddress)
	if err != nil {
		panic(err)
	}
	defer func() { _ = conn.Close() }()

	// write request
	writer := bufio.NewWriter(conn)
	if _, err = writer.Write([]byte{
		0xFF,                                               // start byte
		0x00, 0x02,                                         // protocol version (2)
		0x00,                                               // mode (0)
		byte(ServersCount >> 8), byte(ServersCount & 0xFF), // ip addresses count
		127, 0, 0, 1, byte(10000 >> 8), byte(10000 & 0xFF), // first server (ip + port)
		127, 0, 0, 1, byte(20000 >> 8), byte(20000 & 0xFF), // second server (ip + port)
	}); err != nil {
		panic(err)
	}
	_ = writer.Flush()

	// read response
	reader := bufio.NewReader(conn)
	for i := 0; i < ServersCount; i++ {
		var addr = make([]byte, 6)
		if _, err = io.ReadFull(reader, addr); err != nil {
			panic(err)
		}

		var hostname, mapName, tags string
		var players, maxPlayers, queue uint16
		var wiped uint64

		hostname, _ = reader.ReadString(0)
		mapName, _ = reader.ReadString(0)
		_ = binary.Read(reader, binary.BigEndian, &players)
		_ = binary.Read(reader, binary.BigEndian, &maxPlayers)
		_ = binary.Read(reader, binary.BigEndian, &wiped)
		_ = binary.Read(reader, binary.BigEndian, &queue)
		tags, _ = reader.ReadString(0)

		fmt.Printf("\nResponse for %d.%d.%d.%d:%d\n", addr[0], addr[1], addr[2], addr[3], uint16(addr[4])<<8|uint16(addr[5]))
		fmt.Printf("Hostname: %s\n", hostname)
		fmt.Printf("Map: %s\n", mapName)
		fmt.Printf("Players: %d\n", players)
		fmt.Printf("Max players: %d\n", maxPlayers)
		fmt.Printf("Queue players: %d\n", queue)
		fmt.Printf("Wiped: %s\n", time.Unix(int64(wiped), 0).String())
		fmt.Printf("Tags: %s\n", tags)
	}
}
```