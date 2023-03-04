package RustServer

import (
	"bufio"
	"encoding/binary"
	"rustQueryProxy/Config"
	"strconv"
	"strings"
	"time"
)

type Model struct {
	Address    string
	Hostname   string
	Map        string
	MaxPlayers uint16
	Players    uint16
	Queue      uint16
	Wiped      uint64
	CustomTags string
	UpdateTime time.Time
	QueryPort  uint16
}

func (srv *Model) WriteToStream(writer *bufio.Writer, version uint16) {
	_, _ = writer.WriteString(srv.Hostname)
	_ = writer.WriteByte(0)

	_, _ = writer.WriteString(srv.Map)
	_ = writer.WriteByte(0)

	if version == 1 {
		_ = writer.WriteByte(byte(srv.Players >> 8))
		_ = writer.WriteByte(byte(srv.Players & 0xFF))

		_ = writer.WriteByte(byte(srv.MaxPlayers >> 8))
		_ = writer.WriteByte(byte(srv.MaxPlayers & 0xFF))

		_ = binary.Write(writer, binary.BigEndian, srv.Wiped)
	} else if version == 2 {
		_ = binary.Write(writer, binary.BigEndian, srv.Players)
		_ = binary.Write(writer, binary.BigEndian, srv.MaxPlayers)
		_ = binary.Write(writer, binary.BigEndian, srv.Wiped)
		_ = binary.Write(writer, binary.BigEndian, srv.Queue)

		_, _ = writer.WriteString(srv.CustomTags)
		_ = writer.WriteByte(0)
	}
}

func (srv *Model) Update(data *RawModel) {
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
		for _, v := range Config.CustomTagsWhitelist {
			if strings.EqualFold(tag, v) {
				customTags = append(customTags, tag)
				break
			}
		}
	}
	srv.CustomTags = strings.Join(customTags, ",")

	srv.UpdateTime = time.Now()
}
