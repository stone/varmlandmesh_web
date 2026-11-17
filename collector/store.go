package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Node is everything we know about one MeshCore node.
type Node struct {
	Key      string    `json:"key"`
	Name     string    `json:"name,omitempty"`
	Type     string    `json:"type,omitempty"`
	HasPos   bool      `json:"has_pos,omitempty"`
	Lat      float64   `json:"lat,omitempty"`
	Lon      float64   `json:"lon,omitempty"`
	Last     time.Time `json:"last"`                // last time anything was heard from or about it
	Advert   time.Time `json:"advert,omitzero"`     // when its newest advert was received
	AdvertTS uint32    `json:"advert_ts,omitempty"` // sender clock of that advert
}

// Link is a radio link between two nodes, from a neighbour report.
type Link struct {
	A     string    `json:"a"` // full key, A < B
	B     string    `json:"b"`
	SNR   float64   `json:"snr"`
	Heard time.Time `json:"heard"`
	Count int       `json:"count"`
}

// Store holds all nodes and links in memory and persists them to disk.
type Store struct {
	mu    sync.Mutex
	nodes map[string]*Node
	links map[string]*Link
}

func NewStore() *Store {
	return &Store{nodes: map[string]*Node{}, links: map[string]*Link{}}
}

func (s *Store) node(key string) *Node {
	n := s.nodes[key]
	if n == nil {
		n = &Node{Key: key}
		s.nodes[key] = n
	}
	return n
}

func touch(n *Node, at time.Time) {
	if at.After(n.Last) {
		n.Last = at
	}
}

// Touch records that a node was heard at the given time.
func (s *Store) Touch(key string, at time.Time) {
	key = normKey(key)
	if len(key) != 64 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	touch(s.node(key), at)
}

// AddAdvert stores a verified advert received at the given time.
// Older adverts (replays) only refresh last-heard.
func (s *Store) AddAdvert(a Advert, at time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := s.node(a.Key)
	touch(n, at)
	if a.Timestamp < n.AdvertTS {
		return
	}
	n.AdvertTS = a.Timestamp
	if at.After(n.Advert) {
		n.Advert = at
	}
	n.Type = a.Type
	if a.Name != "" {
		n.Name = a.Name
	}
	if a.HasPos {
		n.HasPos, n.Lat, n.Lon = true, a.Lat, a.Lon
	}
}

// Neighbour is one entry in a neighbour report.
type Neighbour struct {
	Key          string
	SNR          float64
	HeardSecsAgo int
}

// AddNeighbours stores a neighbour report from origin made at time at.
func (s *Store) AddNeighbours(origin string, at time.Time, list []Neighbour) {
	origin = normKey(origin)
	if len(origin) != 64 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	touch(s.node(origin), at)
	for _, nb := range list {
		key := s.resolveKey(nb.Key)
		if key == "" || key == origin {
			continue
		}
		heard := at.Add(-time.Duration(nb.HeardSecsAgo) * time.Second)
		touch(s.node(key), heard)

		a, b := origin, key
		if b < a {
			a, b = b, a
		}
		id := a + "|" + b
		l := s.links[id]
		if l == nil {
			l = &Link{A: a, B: b}
			s.links[id] = l
		}
		l.Count++
		if !heard.Before(l.Heard) {
			l.Heard = heard
			l.SNR = nb.SNR
		}
	}
}

// resolveKey turns a full key or a unique key prefix into a full key.
// Caller holds s.mu.
func (s *Store) resolveKey(k string) string {
	k = normKey(k)
	if len(k) == 64 {
		return k
	}
	if len(k) < 8 {
		return ""
	}
	found := ""
	for full := range s.nodes {
		if strings.HasPrefix(full, k) {
			if found != "" {
				return "" // ambiguous
			}
			found = full
		}
	}
	return found
}

// Prune drops nodes and links not heard since cutoff.
func (s *Store) Prune(cutoff time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, n := range s.nodes {
		if n.Last.Before(cutoff) {
			delete(s.nodes, k)
		}
	}
	for k, l := range s.links {
		if l.Heard.Before(cutoff) {
			delete(s.links, k)
		}
	}
}

// BBox is a lat/lon box: south, north, west, east.
type BBox struct {
	S float64 `json:"s"`
	N float64 `json:"n"`
	W float64 `json:"w"`
	E float64 `json:"e"`
}

func (b BBox) Contains(lat, lon float64) bool {
	return lat >= b.S && lat <= b.N && lon >= b.W && lon <= b.E
}

// Export types: the public nodes.json format.
type ExportNode struct {
	ID     string     `json:"id"` // first 16 hex chars of the key
	Key    string     `json:"key"`
	Name   string     `json:"name"`
	Type   string     `json:"type"`
	Lat    float64    `json:"lat"`
	Lon    float64    `json:"lon"`
	Last   time.Time  `json:"last"`
	Advert *time.Time `json:"advert,omitempty"`
}

type ExportLink struct {
	A     string    `json:"a"`
	B     string    `json:"b"`
	SNR   float64   `json:"snr"`
	Heard time.Time `json:"heard"`
	Count int       `json:"count"`
}

type Export struct {
	Schema    int                       `json:"schema"`
	Generated time.Time                 `json:"generated"`
	Sources   map[string]map[string]any `json:"sources"`
	BBox      BBox                      `json:"bbox"`
	Count     int                       `json:"count"`
	Nodes     []ExportNode              `json:"nodes"`
	Links     []ExportLink              `json:"links"`
}

// Export returns nodes with a position inside bbox heard since cutoff,
// and links between them.
func (s *Store) Export(bbox BBox, cutoff, now time.Time) Export {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := Export{
		Schema:    1,
		Generated: now.UTC(),
		BBox:      bbox,
		Nodes:     []ExportNode{},
		Links:     []ExportLink{},
	}
	ids := map[string]string{}
	for _, n := range s.nodes {
		if !n.HasPos || !bbox.Contains(n.Lat, n.Lon) || n.Last.Before(cutoff) {
			continue
		}
		e := ExportNode{
			ID:   n.Key[:16],
			Key:  n.Key,
			Name: n.Name,
			Type: n.Type,
			Lat:  n.Lat,
			Lon:  n.Lon,
			Last: n.Last.UTC(),
		}
		if e.Name == "" {
			e.Name = e.ID
		}
		if e.Type == "" {
			e.Type = "unknown"
		}
		if !n.Advert.IsZero() {
			t := n.Advert.UTC()
			e.Advert = &t
		}
		ids[n.Key] = e.ID
		out.Nodes = append(out.Nodes, e)
	}
	for _, l := range s.links {
		a, okA := ids[l.A]
		b, okB := ids[l.B]
		if !okA || !okB || l.Heard.Before(cutoff) {
			continue
		}
		out.Links = append(out.Links, ExportLink{A: a, B: b, SNR: l.SNR, Heard: l.Heard.UTC(), Count: l.Count})
	}
	sort.Slice(out.Nodes, func(i, j int) bool { return out.Nodes[i].ID < out.Nodes[j].ID })
	sort.Slice(out.Links, func(i, j int) bool {
		if out.Links[i].A != out.Links[j].A {
			return out.Links[i].A < out.Links[j].A
		}
		return out.Links[i].B < out.Links[j].B
	})
	out.Count = len(out.Nodes)
	out.Sources = map[string]map[string]any{"mqtt": {"ok": true, "count": out.Count}}
	return out
}

type stateFile struct {
	Nodes map[string]*Node `json:"nodes"`
	Links map[string]*Link `json:"links"`
}

// Save writes the store atomically to path.
func (s *Store) Save(path string) error {
	s.mu.Lock()
	data, err := json.Marshal(stateFile{Nodes: s.nodes, Links: s.links})
	s.mu.Unlock()
	if err != nil {
		return err
	}
	return writeFileAtomic(path, data)
}

// writeFileAtomic writes data to a temp file next to path and renames it
// into place, so readers never see a partial file. The file is created
// 0644 from the start (os.CreateTemp would use 0600), so a file watcher
// that opens the temp file does not get "permission denied".
func writeFileAtomic(path string, data []byte) error {
	rnd := make([]byte, 6)
	rand.Read(rnd)
	tmp := filepath.Join(filepath.Dir(path), "."+filepath.Base(path)+"."+hex.EncodeToString(rnd)+".tmp")
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	defer os.Remove(tmp)
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Load reads a store saved by Save. A missing file gives an empty store.
func Load(path string) (*Store, error) {
	s := NewStore()
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	var st stateFile
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, err
	}
	if st.Nodes != nil {
		s.nodes = st.Nodes
	}
	if st.Links != nil {
		s.links = st.Links
	}
	return s, nil
}

func normKey(k string) string {
	return strings.ToLower(strings.TrimSpace(k))
}
