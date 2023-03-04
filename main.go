package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"rustQueryProxy/Config"
	"rustQueryProxy/RustServer"
	"rustQueryProxy/SourceQuery"
	"rustQueryProxy/WebQuery"
	"strings"
	"sync"
	"time"
)

const (
	ProxyVersionMin      = 1
	ProxyVersionMax      = 2
	ProxyModeByteAddress = 0                // Mode: receive addresses as array of bytes
	NetReadTimeout       = 10 * time.Second // Read timeout
)

var cache = make(map[string]*RustServer.Model)
var cacheLock sync.Mutex

func main() {
	Config.Load()
	WebQuery.LoadCache()

	listener, err := net.Listen("tcp", Config.Bind)
	if err != nil {
		log.Fatalf("Listen error: %s\n", err)
	}

	go CacheCleanupWorker()

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
		for _, v := range Config.IpWhitelist {
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
		log.Printf("InConn | %s | Header read error: %s\n", conn.RemoteAddr().String(), err)
		return
	}

	if buf[0] != 0xFF {
		log.Printf("InConn | %s | First byte error: 0x%X", conn.RemoteAddr().String(), buf[0])
		return
	}

	version := uint16(buf[1])<<8 | uint16(buf[2])
	if version < ProxyVersionMin || version > ProxyVersionMax {
		log.Printf("InConn | %s | Version error: %d (min: %d, max: %d)\n", conn.RemoteAddr().String(), version, ProxyVersionMin, ProxyVersionMax)
		return
	}

	mode := buf[3]
	if mode != ProxyModeByteAddress {
		log.Printf("InConn | %s | Mode error: %d\n", conn.RemoteAddr().String(), mode)
		return
	}

	count := uint16(buf[4])<<8 | uint16(buf[5])

	servers := make([]*RustServer.Model, count)

	for i := uint16(0); i < count; i++ {
		// Address: ip(4) + port(2)
		if _, err = io.ReadFull(conn, buf); err != nil {
			log.Printf("InConn | %s | Address read error %d: %s\n", conn.RemoteAddr().String(), i, err)
			return
		}

		address := fmt.Sprintf("%d.%d.%d.%d:%d", buf[0], buf[1], buf[2], buf[3], uint16(buf[4])<<8|uint16(buf[5]))

		srv, ok := cache[address]
		if !ok {
			srv = &RustServer.Model{Address: address}
			cache[address] = srv
		}

		servers[i] = srv
	}

	writer := bufio.NewWriter(conn)
	now := time.Now()

	wg := &sync.WaitGroup{}
	sp := make(chan struct{}, Config.UpdateBurstLimit)

	for _, v := range servers {
		if now.Sub(v.UpdateTime) > Config.QueryConnectTimeout {
			sp <- struct{}{}
			wg.Add(1)
			go func(s *RustServer.Model) {
				UpdateRustServer(s)
				<-sp
				wg.Done()
			}(v)
		}
	}

	wg.Wait()

	for _, v := range servers {
		IpAddressToBytes(v.Address, buf)
		_, _ = writer.Write(buf)

		v.WriteToStream(writer, version)
	}

	if err = writer.Flush(); err != nil {
		log.Printf("InConn | %s | Flush: %s\n",
			conn.RemoteAddr().String(), err)
	}
}

func UpdateRustServer(server *RustServer.Model) {
	var data *RustServer.RawModel
	var err error

	if Config.SteamApiToken != "" {
		data, _ = WebQuery.Query(server.Address)
	}

	if data == nil {
		data, err = SourceQuery.Query(server.Address)
		if err != nil {
			log.Printf("Query | %s | %s\n", server.Address, err.Error())
			return
		}
	}

	server.Update(data)
}

func CacheCleanupWorker() {
	for {
		time.Sleep(Config.ServerCacheTime)

		cacheLock.Lock()

		old := make([]string, 0)
		now := time.Now()

		for k, v := range cache {
			if now.Sub(v.UpdateTime) > Config.ServerCacheTime {
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
