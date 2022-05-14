package main

import (
	"bufio"
	"errors"
	"io"
	"net"
	"strings"
	"time"
)

type RawRustServerModel struct {
	Hostname string
	Map      string
	Tags     []string
}

func QueryRustServer(address string) (*RawRustServerModel, error) {
	model := RawRustServerModel{}

	challenge := make([]byte, 0)

	var r *bufio.Reader
	var conn net.Conn
	var err error

	for {
		conn, err = net.DialTimeout("udp", address, cfg.QueryConnectTimeoutInSeconds*time.Second)
		if err != nil {
			return nil, err
		}

		_ = conn.SetDeadline(time.Now().Add(cfg.QueryConnectTimeoutInSeconds * time.Second))

		_, err = conn.Write(append([]byte("\xFF\xFF\xFF\xFFTSource Engine Query\x00"), challenge...))
		if err != nil {
			_ = conn.Close()
			return nil, err
		}

		r = bufio.NewReader(conn)

		_, _ = r.Discard(4) // header (FF FF FF FF)

		if x, _ := r.ReadByte(); x == '\x41' {
			if len(challenge) != 0 {
				_ = conn.Close()
				return nil, errors.New("challenge failed")
			}
			challenge = make([]byte, 4)
			_, _ = io.ReadFull(r, challenge)
			_ = conn.Close()
			continue
		}

		_, _ = r.Discard(1) // protocol
		break
	}

	model.Hostname, _ = r.ReadString(0)             // name
	model.Map, _ = r.ReadString(0)                  // map
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
		keywords = strings.TrimRight(keywords, "\u0000")
		keywords = strings.ToLower(keywords)
		model.Tags = strings.Split(keywords, ",")
	}

	if edf&0x01 != 0 {
		_, _ = r.Discard(8) // GameID
	}

	_ = conn.Close()

	model.Hostname = strings.TrimRight(model.Hostname, "\u0000")
	model.Map = strings.TrimRight(model.Map, "\u0000")

	if model.Hostname == "" {
		return nil, errors.New("hostname is empty")
	}

	if model.Map == "" {
		return nil, errors.New("map is empty")
	}

	if len(model.Tags) == 0 {
		return nil, errors.New("tags is empty")
	}

	return &model, nil
}
