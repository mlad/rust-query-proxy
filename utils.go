package main

import "strconv"

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
