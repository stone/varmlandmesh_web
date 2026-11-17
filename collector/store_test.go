package main

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var varmland = BBox{S: 58.8, N: 61.1, W: 11.6, E: 14.5}

func key(c string) string { return strings.Repeat(c, 64) }

func TestExportFiltersByBBoxPositionAndAge(t *testing.T) {
	now := time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)
	s := NewStore()
	s.AddAdvert(Advert{Key: key("a"), Timestamp: 10, Type: "repeater", Name: "Karlstad", HasPos: true, Lat: 59.38, Lon: 13.5}, now)
	s.AddAdvert(Advert{Key: key("b"), Timestamp: 10, Type: "repeater", Name: "Borlänge", HasPos: true, Lat: 60.48, Lon: 15.43}, now)
	s.AddAdvert(Advert{Key: key("c"), Timestamp: 10, Type: "companion", Name: "Utan pos"}, now)
	s.AddAdvert(Advert{Key: key("d"), Timestamp: 10, Type: "repeater", Name: "Gammal", HasPos: true, Lat: 59.4, Lon: 13.5}, now.Add(-100*24*time.Hour))

	e := s.Export(varmland, now.Add(-90*24*time.Hour), now)
	if e.Count != 1 || e.Nodes[0].Name != "Karlstad" {
		t.Fatalf("got %+v, want only Karlstad", e.Nodes)
	}
	if e.Nodes[0].ID != key("a")[:16] {
		t.Errorf("id = %s", e.Nodes[0].ID)
	}
}

func TestOlderAdvertDoesNotOverwritePosition(t *testing.T) {
	now := time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)
	s := NewStore()
	s.AddAdvert(Advert{Key: key("a"), Timestamp: 200, Type: "repeater", Name: "Ny", HasPos: true, Lat: 59.5, Lon: 13.0}, now)
	s.AddAdvert(Advert{Key: key("a"), Timestamp: 100, Type: "repeater", Name: "Gammal", HasPos: true, Lat: 59.0, Lon: 12.0}, now.Add(time.Minute))

	n := s.nodes[key("a")]
	if n.Name != "Ny" || n.Lat != 59.5 {
		t.Errorf("replayed advert overwrote node: %+v", n)
	}
	if !n.Last.Equal(now.Add(time.Minute)) {
		t.Errorf("last = %v, want refreshed by replay", n.Last)
	}
}

func TestNeighboursCreateLinksBetweenExportedNodes(t *testing.T) {
	now := time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)
	s := NewStore()
	s.AddAdvert(Advert{Key: key("a"), Timestamp: 1, Type: "repeater", HasPos: true, Lat: 59.4, Lon: 13.5}, now)
	s.AddAdvert(Advert{Key: key("b"), Timestamp: 1, Type: "repeater", HasPos: true, Lat: 59.9, Lon: 12.3}, now)

	// Reported by b about a (uppercase key, as in MQTT), and by a about b via a prefix.
	s.AddNeighbours(strings.ToUpper(key("b")), now, []Neighbour{{Key: strings.ToUpper(key("a")), SNR: 5.5, HeardSecsAgo: 60}})
	s.AddNeighbours(key("a"), now.Add(time.Minute), []Neighbour{{Key: key("b")[:12], SNR: 7, HeardSecsAgo: 0}})
	// Unknown neighbour outside the export is ignored in output.
	s.AddNeighbours(key("a"), now, []Neighbour{{Key: key("f"), SNR: 1}})

	e := s.Export(varmland, now.Add(-time.Hour), now.Add(time.Hour))
	if len(e.Links) != 1 {
		t.Fatalf("links = %+v, want 1", e.Links)
	}
	l := e.Links[0]
	if l.A != key("a")[:16] || l.B != key("b")[:16] || l.Count != 2 || l.SNR != 7 {
		t.Errorf("link = %+v", l)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	now := time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)
	s := NewStore()
	s.AddAdvert(Advert{Key: key("a"), Timestamp: 1, Type: "room", Name: "Rum", HasPos: true, Lat: 59.4, Lon: 13.5}, now)
	path := filepath.Join(t.TempDir(), "state.json")
	if err := s.Save(path); err != nil {
		t.Fatal(err)
	}
	s2, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if e := s2.Export(varmland, now.Add(-time.Hour), now); e.Count != 1 || e.Nodes[0].Name != "Rum" {
		t.Fatalf("after load: %+v", e.Nodes)
	}
	if _, err := Load(filepath.Join(t.TempDir(), "missing.json")); err != nil {
		t.Errorf("missing state file should give empty store, got %v", err)
	}
}

func TestPrune(t *testing.T) {
	now := time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)
	s := NewStore()
	s.Touch(key("a"), now.Add(-200*24*time.Hour))
	s.Touch(key("b"), now)
	s.Prune(now.Add(-90 * 24 * time.Hour))
	if _, ok := s.nodes[key("a")]; ok {
		t.Error("old node not pruned")
	}
	if _, ok := s.nodes[key("b")]; !ok {
		t.Error("fresh node pruned")
	}
}
