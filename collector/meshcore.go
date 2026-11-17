package main

import (
	"crypto/ed25519"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

// MeshCore payload types we care about.
const payloadTypeAdvert = 0x04

// Advert flag bits (first byte of app data).
const (
	advFlagHasLocation = 0x10
	advFlagFeature1    = 0x20
	advFlagFeature2    = 0x40
	advFlagHasName     = 0x80
)

var nodeTypes = map[byte]string{
	1: "companion",
	2: "repeater",
	3: "room",
	4: "sensor",
}

// Advert is a decoded and signature-checked MeshCore advertisement.
type Advert struct {
	Key       string // full public key, lowercase hex
	Timestamp uint32 // sender clock, unix seconds
	Type      string
	Name      string
	HasPos    bool
	Lat, Lon  float64
}

// splitPacket returns the payload type and payload of a raw MeshCore packet.
//
// Layout: header(1) [transport codes(4)] path_len(1) path(n) payload.
// Header bits 0-1 are the route type, bits 2-5 the payload type.
// path_len bits 0-5 are the hop count, bits 6-7 the hash size minus one.
func splitPacket(raw []byte) (byte, []byte, error) {
	if len(raw) < 2 {
		return 0, nil, errors.New("packet too short")
	}
	header := raw[0]
	i := 1
	if route := header & 0x03; route == 0x00 || route == 0x03 {
		i += 4 // transport codes
	}
	if i >= len(raw) {
		return 0, nil, errors.New("packet too short for path")
	}
	pathLen := raw[i]
	i++
	i += int(pathLen&0x3F) * (int(pathLen>>6) + 1)
	if i > len(raw) {
		return 0, nil, errors.New("path longer than packet")
	}
	return (header >> 2) & 0x0F, raw[i:], nil
}

// decodeAdvert parses an advert payload and verifies its Ed25519 signature.
//
// Payload: pubkey(32) timestamp(4, LE) signature(64) appdata.
// The signature covers pubkey + timestamp + appdata.
func decodeAdvert(p []byte) (Advert, error) {
	const head = 32 + 4 + 64
	if len(p) < head+1 {
		return Advert{}, fmt.Errorf("advert too short: %d bytes", len(p))
	}
	pub := p[:32]
	ts := p[32:36]
	sig := p[36:100]
	app := p[100:]

	msg := make([]byte, 0, 36+len(app))
	msg = append(msg, pub...)
	msg = append(msg, ts...)
	msg = append(msg, app...)
	if !ed25519.Verify(ed25519.PublicKey(pub), msg, sig) {
		return Advert{}, errors.New("bad advert signature")
	}

	a := Advert{
		Key:       fmt.Sprintf("%x", pub),
		Timestamp: binary.LittleEndian.Uint32(ts),
	}
	flags := app[0]
	a.Type = nodeTypes[flags&0x0F]
	if a.Type == "" {
		a.Type = "unknown"
	}
	j := 1
	if flags&advFlagHasLocation != 0 {
		if len(app) < j+8 {
			return Advert{}, errors.New("advert location truncated")
		}
		lat := int32(binary.LittleEndian.Uint32(app[j:]))
		lon := int32(binary.LittleEndian.Uint32(app[j+4:]))
		j += 8
		if lat != 0 || lon != 0 {
			a.HasPos = true
			a.Lat = float64(lat) / 1e6
			a.Lon = float64(lon) / 1e6
		}
	}
	if flags&advFlagFeature1 != 0 {
		j += 2
	}
	if flags&advFlagFeature2 != 0 {
		j += 2
	}
	if flags&advFlagHasName != 0 && j < len(app) {
		a.Name = strings.TrimRight(strings.ToValidUTF8(string(app[j:]), ""), "\x00")
	}
	return a, nil
}
