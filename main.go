package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	ProxyVersion         = 2                // Proxy-server version
	ProxyModeByteAddress = 0                // Mode: receive addresses as array of bytes
	NetReadTimeout       = 10 * time.Second // Read timeout
)

var cache = make(map[string]*RustServerInfo)
var cacheLock sync.Mutex

func main() {
	LoadConfig()

	listener, err := net.Listen("tcp", cfg.Bind)
	if err != nil {
		log.Fatalf("Listen error: %s\n", err)
	}

	go CleanupCache()

	log.Printf("Server started: %s\n", listener.Addr().String())

	var conn net.Conn
	for {
		conn, err = listener.Accept()
		if err != nil {
			log.Printf("Incoming connection: %s\n", err)
			continue
		}

		address := conn.RemoteAddr().String()
		address = address[:strings.LastIndexByte(address, ':')]
		found := false
		for _, v := range cfg.IpWhitelist {
			if address == v {
				found = true
				break
			}
		}

		if !found {
			log.Printf("Not found in whitelist: %s\n", address)
			_ = conn.Close()
			continue
		}

		cacheLock.Lock()
		IncomingConnection(conn)
		cacheLock.Unlock()

		_ = conn.Close()
	}
}

func IncomingConnection(conn net.Conn) {
	buf := make([]byte, 6)
	var err error

	_ = conn.SetReadDeadline(time.Now().Add(NetReadTimeout))

	// Header: 0xFF + ver(2) + mode(1) + count(2)
	if _, err = io.ReadFull(conn, buf); err != nil {
		log.Printf("InConn | %s | Header read error: %s\n",
			conn.RemoteAddr().String(), err)
		return
	}

	if buf[0] != 0xFF {
		log.Printf("InConn | %s | First byte error: 0x%X",
			conn.RemoteAddr().String(), buf[0])
		return
	}

	version := uint16(buf[1])<<8 | uint16(buf[2])
	if version != ProxyVersion {
		log.Printf("InConn | %s | Version error: %d (server: %d)\n",
			conn.RemoteAddr().String(), version, ProxyVersion)
		return
	}

	mode := buf[3]
	if mode != ProxyModeByteAddress {
		log.Printf("InConn | %s | Mode error: %d\n",
			conn.RemoteAddr().String(), mode)
		return
	}

	count := uint16(buf[4])<<8 | uint16(buf[5])

	servers := make([]*RustServerInfo, count)

	for i := uint16(0); i < count; i++ {
		// Address: ip(4) + port(2)
		if _, err = io.ReadFull(conn, buf); err != nil {
			log.Printf("InConn | %s | Address read error %d: %s\n",
				conn.RemoteAddr().String(), i, err)
			return
		}

		address := fmt.Sprintf("%d.%d.%d.%d:%d", buf[0], buf[1], buf[2], buf[3], uint16(buf[4])<<8|uint16(buf[5]))

		srv, ok := cache[address]
		if !ok {
			srv = &RustServerInfo{Address: address}
			cache[address] = srv
		}

		servers[i] = srv
	}

	writer := bufio.NewWriter(conn)
	now := time.Now()

	wg := &sync.WaitGroup{}
	sp := make(chan struct{}, cfg.UpdateBurstLimit)

	for _, v := range servers {
		if now.Sub(v.UpdateTime) > cfg.QueryConnectTimeoutInSeconds*time.Second {
			sp <- struct{}{}
			wg.Add(1)
			go v.UpdateInfoAsync(sp, wg)
		}
	}

	wg.Wait()

	for _, v := range servers {
		IpAddressToBytes(v.Address, buf)
		_, _ = writer.Write(buf)

		v.WriteQueryProxy(writer)
	}

	if err = writer.Flush(); err != nil {
		log.Printf("InConn | %s | Flush: %s\n",
			conn.RemoteAddr().String(), err)
	}
}

func CleanupCache() {
	for {
		time.Sleep(cfg.ServerCacheTimeInSeconds * time.Second)

		cacheLock.Lock()

		old := make([]string, 0)
		now := time.Now()

		for k, v := range cache {
			if now.Sub(v.UpdateTime) > cfg.ServerCacheTimeInSeconds*time.Second {
				old = append(old, k)
			}
		}

		for _, v := range old {
			delete(cache, v)
		}

		cacheLock.Unlock()

		old = nil
	}
}

func IpAddressToBytes(address string, buf []byte) {
	bufPos := 0
	addrPos := 0
	var tmp uint64

	for k, v := range address {
		if v == '.' || v == ':' {
			tmp, _ = strconv.ParseUint(address[addrPos:k], 10, 8)

			buf[bufPos] = byte(tmp)
			bufPos++
			addrPos = k + 1

			if v == ':' {
				tmp, _ = strconv.ParseUint(address[k+1:], 10, 16)
				buf[bufPos] = byte(tmp >> 8)
				buf[bufPos+1] = byte(tmp & 0xFF)
				break
			}
		}
	}
}
