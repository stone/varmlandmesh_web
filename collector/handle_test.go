package main

import (
	"testing"
	"time"
)

func TestHandleMessages(t *testing.T) {
	now := time.Date(2026, 11, 17, 17, 0, 0, 0, time.UTC)
	s := NewStore()
	st := &stats{}

	// Advert from SE1880-EYRASEC (Örebro), widened bbox below so it is exported.
	handle(s, st, "meshcore/ORB/A1/packets", []byte(`{"timestamp":"2026-11-17T16:42:47.090224+00:00","type":"PACKET","packet_type":"4","raw":"`+advertFixtures[2].raw+`","origin_id":"`+key("e")+`"}`), now)
	// Neighbour report from the observer about the advertiser.
	handle(s, st, "meshcore/ORB/A1/neighbors", []byte(`{"timestamp":"2026-11-17T16:50:00+00:00","origin_id":"`+key("e")+`","neighbors":[{"pubkey":"`+"A558CE1A45481CE4"+`","snr":6.5,"heard_secs_ago":30}]}`), now)
	// Status without zone, from a node that is not otherwise known.
	handle(s, st, "meshcore/KSD/B2/status", []byte(`{"status":"online","timestamp":"2026-11-17T16:55:00.5","origin_id":"`+key("9")+`"}`), now)
	// Garbage is ignored.
	handle(s, st, "meshcore/KSD/B2/packets", []byte(`not json`), now)

	if st.adverts.Load() != 1 || st.badAdverts.Load() != 0 {
		t.Fatalf("adverts=%d bad=%d", st.adverts.Load(), st.badAdverts.Load())
	}
	if n := s.nodes[key("9")]; n == nil || !n.Last.Equal(time.Date(2026, 11, 17, 16, 55, 0, 5e8, time.UTC)) {
		t.Errorf("status node = %+v", n)
	}

	// The observer has no position, so give it one to see the link exported.
	s.AddAdvert(Advert{Key: key("e"), Timestamp: 1, Type: "repeater", HasPos: true, Lat: 59.3, Lon: 15.2}, now)
	e := s.Export(BBox{S: 59, N: 60, W: 15, E: 16}, now.Add(-time.Hour), now)
	if e.Count != 2 || len(e.Links) != 1 || e.Links[0].SNR != 6.5 {
		t.Fatalf("export = %+v", e)
	}
}

func TestParseTime(t *testing.T) {
	now := time.Date(2026, 11, 17, 17, 0, 0, 0, time.UTC)
	if got := parseTime("", now); !got.Equal(now) {
		t.Errorf("empty: %v", got)
	}
	if got := parseTime("2030-01-01T00:00:00Z", now); !got.Equal(now) {
		t.Errorf("future: %v", got)
	}
	if got := parseTime("2026-11-17T10:00:00+01:00", now); !got.Equal(time.Date(2026, 11, 17, 9, 0, 0, 0, time.UTC)) {
		t.Errorf("zoned: %v", got)
	}
}
