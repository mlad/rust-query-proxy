package main

import (
	"bufio"
	"encoding/binary"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	UpdateStatusOk = iota
	UpdateStatusConnErr
	UpdateStatusWriteErr
)

type RustServerInfo struct {
	Address      string
	Hostname     string
	Map          string
	MaxPlayers   int
	Players      int
	Wiped        int64
	UpdateStatus byte
	UpdateTime   time.Time
}

func (srv *RustServerInfo) WriteQueryProxy(writer *bufio.Writer) {
	_, _ = writer.WriteString(srv.Hostname)
	_ = writer.WriteByte(0)

	_, _ = writer.WriteString(srv.Map)
	_ = writer.WriteByte(0)

	_ = writer.WriteByte(byte(srv.Players >> 8))
	_ = writer.WriteByte(byte(srv.Players & 0xFF))

	_ = writer.WriteByte(byte(srv.MaxPlayers >> 8))
	_ = writer.WriteByte(byte(srv.MaxPlayers & 0xFF))

	_ = binary.Write(writer, binary.BigEndian, srv.Wiped)
}

func (srv *RustServerInfo) UpdateInfoAsync(sp <-chan struct{}, wg *sync.WaitGroup) {
	srv.UpdateInfo()
	<-sp
	wg.Done()
}

func (srv *RustServerInfo) UpdateInfo() {
	conn, err := net.DialTimeout("udp", srv.Address, cfg.QueryConnectTimeout)
	if err != nil {
		log.Printf("UpdInf | %s | Dial: %s\n", srv.Address, err)
		srv.UpdateStatus = UpdateStatusConnErr
		return
	}

	_ = conn.SetDeadline(time.Now().Add(cfg.QueryConnectTimeout))

	_, err = conn.Write([]byte("\xFF\xFF\xFF\xFFTSource Engine Query\x00"))
	if err != nil {
		log.Printf("UpdInf | %s | Write: %s\n", srv.Address, err)
		srv.UpdateStatus = UpdateStatusWriteErr
		_ = conn.Close()
		return
	}

	r := bufio.NewReader(conn)

	_, _ = r.Discard(4 + 1 + 1)                     // header + protocol
	srv.Hostname, _ = r.ReadString(0)               // name
	srv.Map, _ = r.ReadString(0)                    // map
	_, _ = r.ReadString(0)                          // folder
	_, _ = r.ReadString(0)                          // game
	_, _ = r.Discard(2 + 1 + 1 + 1 + 1 + 1 + 1 + 1) // app id + Players + MaxPlayers + Bots + Dedicated + Os + Password + Secure
	_, _ = r.ReadString(0)

	edf, _ := r.ReadByte()
	if edf&0x80 != 0 {
		_, _ = r.Discard(2) // Game port
	}

	if edf&0x10 != 0 {
		_, _ = r.Discard(8) // SteamID
	}

	if edf&0x40 != 0 {
		_, _ = r.Discard(2)    // SpecPort
		_, _ = r.ReadString(0) // SpecName
	}

	if edf&0x20 != 0 {
		keywords, _ := r.ReadString(0) // Keywords

		for _, i := range strings.Split(keywords, ",") {
			switch {
			case strings.HasPrefix(i, "mp"):
				srv.MaxPlayers, _ = strconv.Atoi(i[2:])
			case strings.HasPrefix(i, "cp"):
				srv.Players, _ = strconv.Atoi(i[2:])
			case strings.HasPrefix(i, "born"):
				srv.Wiped, _ = strconv.ParseInt(i[4:], 10, 64)
			}
		}
	}

	if edf&0x01 != 0 {
		_, _ = r.Discard(8) // GameID
	}

	_ = conn.Close()

	if srv.Hostname != "" {
		srv.Hostname = srv.Hostname[:len(srv.Hostname)-1]
	}

	if srv.Map != "" {
		srv.Map = srv.Map[:len(srv.Map)-1]
	}

	srv.UpdateTime = time.Now()
	srv.UpdateStatus = UpdateStatusOk
}
