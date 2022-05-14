package main

import (
	"bufio"
	"encoding/binary"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RustServerInfo struct {
	Address    string
	Hostname   string
	Map        string
	MaxPlayers uint16
	Players    uint16
	Queue      uint16
	Wiped      uint64
	CustomTags string
	UpdateTime time.Time
}

func (srv *RustServerInfo) WriteQueryProxy(writer *bufio.Writer) {
	_, _ = writer.WriteString(srv.Hostname)
	_ = writer.WriteByte(0)

	_, _ = writer.WriteString(srv.Map)
	_ = writer.WriteByte(0)

	// _ = writer.WriteByte(byte(srv.Players >> 8))
	// _ = writer.WriteByte(byte(srv.Players & 0xFF))
	//
	// _ = writer.WriteByte(byte(srv.MaxPlayers >> 8))
	// _ = writer.WriteByte(byte(srv.MaxPlayers & 0xFF))

	_ = binary.Write(writer, binary.BigEndian, srv.Players)
	_ = binary.Write(writer, binary.BigEndian, srv.MaxPlayers)
	_ = binary.Write(writer, binary.BigEndian, srv.Wiped)
	_ = binary.Write(writer, binary.BigEndian, srv.Queue)

	_, _ = writer.WriteString(srv.CustomTags)
	_ = writer.WriteByte(0)
}

func (srv *RustServerInfo) UpdateInfoAsync(sp <-chan struct{}, wg *sync.WaitGroup) {
	srv.UpdateInfo()
	<-sp
	wg.Done()
}

func (srv *RustServerInfo) UpdateInfo() {
	data, err := QueryRustServer(srv.Address)
	if err != nil {
		log.Printf("QueryRustServer | %s | %s\n", srv.Address, err.Error())
		return
	}

	srv.Hostname = data.Hostname
	srv.Map = data.Map

	srv.MaxPlayers = 0
	srv.Players = 0
	srv.Queue = 0
	srv.Wiped = 0

	for _, v := range data.Tags {
		switch {
		case strings.HasPrefix(v, "mp"):
			maxPlayers, _ := strconv.ParseUint(v[2:], 10, 16)
			srv.MaxPlayers = uint16(maxPlayers)
		case strings.HasPrefix(v, "cp"):
			currentPlayers, _ := strconv.ParseUint(v[2:], 10, 16)
			srv.Players = uint16(currentPlayers)
		case strings.HasPrefix(v, "qp"):
			queuedPlayers, _ := strconv.ParseUint(v[2:], 10, 16)
			srv.Queue = uint16(queuedPlayers)
		case strings.HasPrefix(v, "born"):
			wipeTimestamp, _ := strconv.ParseUint(v[4:], 10, 64)
			srv.Wiped = wipeTimestamp
		}
	}

	customTags := make([]string, 0)
	for _, tag := range data.Tags {
		for _, v := range cfg.CustomTagsWhitelist {
			if strings.EqualFold(tag, v) {
				customTags = append(customTags, tag)
				break
			}
		}
	}
	srv.CustomTags = strings.Join(customTags, ",")

	srv.UpdateTime = time.Now()
}
