package main

import (
	"encoding/hex"
	"math"
	"testing"
)

// Real adverts captured from mqtt.meshat.se on 2026-11-17.
var advertFixtures = []struct {
	name    string
	raw     string
	keyPfx  string
	advName string
	typ     string
	lat     float64
	lon     float64
}{
	{
		name:    "flood, 1 hop, 2-byte hashes",
		raw:     "1141EBD6EA3350B136D9869AF81A689E02F2C20704DF4F24331CE19FB270C7E2250B80F86784FC6A40177FBD3B45F6682905F6C414EB009497473065C0239C5C3FA783A72C071BCC5D35DFCC3D52F5CBB09650074D5189BD4471E6ABD95621D0C521E216B5C4D6079260CA9A0304E7EB0053452D424C45204D6AC3A46C6761",
		keyPfx:  "ea3350b136d9869a",
		advName: "SE-BLE Mjälga",
		typ:     "repeater",
		lat:     60.476,
		lon:     15.4601,
	},
	{
		name:    "transport flood, 2 hops, 3-byte hashes",
		raw:     "10E1A7000082620CD47ABEC437900FE18624A62C6BEAF25C3502B8C022F083AC29E8D5F03CB07E4BABABFD73D1DC4866EC064F0CEC34CE9FA350F8091EEE6A71BDF679780D8FD910D83ADD58FB64F714075F4450046F4E7038A749E8430248A25FC7BE053D7CD8D59D121D8D7A654E0A9286655703351CDE0053452D313038332D5362672D4C424F",
		keyPfx:  "37900fe18624a62c",
		advName: "SE-1083-Sbg-LBO",
		typ:     "repeater",
		lat:     56.05927,
		lon:     14.556213,
	},
	{
		name:    "flood, 0 hops",
		raw:     "1180A558CE1A45481CE484DD493ABCC6DC6A544F2BCA7BD73709D54FC652852CBAEBD884FC6ACFFB7C290F9EAF8CB57D34153A2A85774C1D55ECF64FD9A6DC4C479C57B511A1583B25CFED371AFC1D0D74BD99F167B4D40E4767C6CCA7100EB1930156FC040592B5A88803A435E8005345313838302D45595241534543",
		keyPfx:  "a558ce1a45481ce4",
		advName: "SE1880-EYRASEC",
		typ:     "repeater",
		lat:     59.287733,
		lon:     15.218084,
	},
}

func TestDecodeAdvertFixtures(t *testing.T) {
	for _, f := range advertFixtures {
		t.Run(f.name, func(t *testing.T) {
			raw, err := hex.DecodeString(f.raw)
			if err != nil {
				t.Fatal(err)
			}
			pt, payload, err := splitPacket(raw)
			if err != nil {
				t.Fatal(err)
			}
			if pt != payloadTypeAdvert {
				t.Fatalf("payload type = %d, want %d", pt, payloadTypeAdvert)
			}
			a, err := decodeAdvert(payload)
			if err != nil {
				t.Fatal(err)
			}
			if a.Key[:16] != f.keyPfx {
				t.Errorf("key = %s, want prefix %s", a.Key, f.keyPfx)
			}
			if a.Name != f.advName {
				t.Errorf("name = %q, want %q", a.Name, f.advName)
			}
			if a.Type != f.typ {
				t.Errorf("type = %q, want %q", a.Type, f.typ)
			}
			if !a.HasPos || math.Abs(a.Lat-f.lat) > 1e-6 || math.Abs(a.Lon-f.lon) > 1e-6 {
				t.Errorf("pos = %v %f,%f, want %f,%f", a.HasPos, a.Lat, a.Lon, f.lat, f.lon)
			}
		})
	}
}

func TestDecodeAdvertRejectsTamperedName(t *testing.T) {
	raw, _ := hex.DecodeString(advertFixtures[0].raw)
	raw[len(raw)-1] ^= 0x01 // flip a bit in the name
	_, payload, err := splitPacket(raw)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := decodeAdvert(payload); err == nil {
		t.Fatal("expected signature error for tampered advert")
	}
}

func TestSplitPacketShort(t *testing.T) {
	for _, raw := range [][]byte{nil, {0x11}, {0x10, 0, 0}, {0x11, 0x05, 0x01}} {
		if _, _, err := splitPacket(raw); err == nil {
			t.Errorf("splitPacket(%x): expected error", raw)
		}
	}
}
