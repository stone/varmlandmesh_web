// Command collector listens to MeshCore traffic on MQTT, keeps a list of
// nodes and radio links, and publishes the nodes inside a bounding box as
// nodes.json to a GitHub branch.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type config struct {
	Broker   string
	User     string
	Pass     string
	Topics   []string
	BBox     BBox
	MaxAge   time.Duration
	Interval time.Duration
	State    string
	Output   string
	GitHub   *GitHub
	DataPath string
}

func env(k, def string) string {
	if v, ok := os.LookupEnv(k); ok && v != "" {
		return v
	}
	return def
}

func loadConfig() (config, error) {
	c := config{
		Broker:   env("MQTT_BROKER", "tcp://mqtt.meshat.se:1883"),
		User:     env("MQTT_USERNAME", "msh"),
		Pass:     env("MQTT_PASSWORD", "msh"),
		Topics:   strings.Split(env("MQTT_TOPICS", "meshcore/+/+/packets,meshcore/+/+/neighbors,meshcore/+/+/status"), ","),
		State:    env("STATE_FILE", "/data/state.json"),
		Output:   env("OUTPUT_FILE", ""),
		DataPath: env("DATA_PATH", "nodes.json"),
	}
	var err error
	if c.MaxAge, err = time.ParseDuration(env("MAX_AGE", "2160h")); err != nil {
		return c, fmt.Errorf("MAX_AGE: %w", err)
	}
	if c.Interval, err = time.ParseDuration(env("PUBLISH_INTERVAL", "15m")); err != nil {
		return c, fmt.Errorf("PUBLISH_INTERVAL: %w", err)
	}
	parts := strings.Split(env("BBOX", "58.8,61.1,11.6,14.5"), ",")
	if len(parts) != 4 {
		return c, errors.New("BBOX must be south,north,west,east")
	}
	var f [4]float64
	for i, p := range parts {
		if f[i], err = strconv.ParseFloat(strings.TrimSpace(p), 64); err != nil {
			return c, fmt.Errorf("BBOX: %w", err)
		}
	}
	c.BBox = BBox{S: f[0], N: f[1], W: f[2], E: f[3]}

	if tok := os.Getenv("GITHUB_TOKEN"); tok != "" {
		c.GitHub = &GitHub{
			API:    env("GITHUB_API", "https://api.github.com"),
			Repo:   env("GITHUB_REPO", "stone/varmlandmesh_web"),
			Branch: env("GITHUB_BRANCH", "data"),
			Token:  tok,
		}
	}
	return c, nil
}

// Counters for logging.
type stats struct {
	messages, adverts, badAdverts atomic.Int64
}

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	if cfg.GitHub == nil && cfg.Output == "" {
		slog.Warn("neither GITHUB_TOKEN nor OUTPUT_FILE set; nodes.json will not be written anywhere")
	}

	store, err := Load(cfg.State)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}
	slog.Info("state loaded", "file", cfg.State, "nodes", len(store.nodes), "links", len(store.links))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	st := &stats{}
	client := connect(cfg, store, st)
	defer client.Disconnect(1000)

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()
	var lastMessages int64
	for {
		select {
		case <-ctx.Done():
			slog.Info("shutting down")
			return store.Save(cfg.State)
		case <-ticker.C:
			msgs := st.messages.Load()
			fresh := msgs != lastMessages
			lastMessages = msgs
			if err := flush(ctx, cfg, store, st, fresh); err != nil {
				slog.Error("flush", "err", err)
			}
		}
	}
}

// flush prunes, saves state and publishes nodes.json.
// It skips publishing when no MQTT messages arrived since last time, so the
// "generated" time in nodes.json shows when data was last live.
func flush(ctx context.Context, cfg config, store *Store, st *stats, fresh bool) error {
	now := time.Now()
	cutoff := now.Add(-cfg.MaxAge)
	store.Prune(cutoff)
	if err := store.Save(cfg.State); err != nil {
		return fmt.Errorf("save state: %w", err)
	}
	exp := store.Export(cfg.BBox, cutoff, now)
	slog.Info("export", "nodes", exp.Count, "links", len(exp.Links),
		"messages", st.messages.Load(), "adverts", st.adverts.Load(), "bad_adverts", st.badAdverts.Load())

	if !fresh {
		slog.Warn("no MQTT messages since last publish; skipping")
		return nil
	}
	data, err := json.MarshalIndent(exp, "", " ")
	if err != nil {
		return err
	}
	if cfg.Output != "" {
		if err := writeFileAtomic(cfg.Output, data); err != nil {
			return fmt.Errorf("write %s: %w", cfg.Output, err)
		}
	}
	if cfg.GitHub != nil && exp.Count == 0 {
		// Never replace published data with an empty list, e.g. after losing state.
		slog.Info("no nodes in bbox yet; skipping GitHub publish")
		return nil
	}
	if cfg.GitHub != nil {
		pctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
		msg := fmt.Sprintf("data: %d noder, %d länkar", exp.Count, len(exp.Links))
		sha, err := cfg.GitHub.Publish(pctx, map[string][]byte{cfg.DataPath: data}, msg)
		if err != nil {
			return fmt.Errorf("publish: %w", err)
		}
		slog.Info("published", "repo", cfg.GitHub.Repo, "branch", cfg.GitHub.Branch, "commit", sha)
	}
	return nil
}

func connect(cfg config, store *Store, st *stats) mqtt.Client {
	id := make([]byte, 4)
	rand.Read(id)
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.Broker).
		SetUsername(cfg.User).
		SetPassword(cfg.Pass).
		SetClientID("varmlandmesh-collector-" + hex.EncodeToString(id)).
		SetCleanSession(true).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetMaxReconnectInterval(time.Minute).
		SetConnectionLostHandler(func(_ mqtt.Client, err error) {
			slog.Warn("mqtt connection lost", "err", err)
		}).
		SetOnConnectHandler(func(c mqtt.Client) {
			slog.Info("mqtt connected", "broker", cfg.Broker)
			filters := map[string]byte{}
			for _, t := range cfg.Topics {
				filters[strings.TrimSpace(t)] = 0
			}
			tok := c.SubscribeMultiple(filters, func(_ mqtt.Client, m mqtt.Message) {
				st.messages.Add(1)
				handle(store, st, m.Topic(), m.Payload(), time.Now())
			})
			if tok.WaitTimeout(30*time.Second) && tok.Error() != nil {
				slog.Error("mqtt subscribe", "err", tok.Error())
			}
		})
	c := mqtt.NewClient(opts)
	c.Connect() // retries in the background
	return c
}

// handle processes one MQTT message. Topic: meshcore/<region>/<observer key>/<kind>.
func handle(store *Store, st *stats, topic string, payload []byte, now time.Time) {
	var msg struct {
		Timestamp string `json:"timestamp"`
		OriginID  string `json:"origin_id"`
		Status    string `json:"status"`
		Raw       string `json:"raw"`
		Neighbors []struct {
			Pubkey       string  `json:"pubkey"`
			SNR          float64 `json:"snr"`
			HeardSecsAgo int     `json:"heard_secs_ago"`
		} `json:"neighbors"`
	}
	if err := json.Unmarshal(payload, &msg); err != nil {
		return
	}
	at := parseTime(msg.Timestamp, now)

	switch path.Base(topic) {
	case "status":
		if msg.Status == "online" {
			store.Touch(msg.OriginID, at)
		}
	case "neighbors":
		list := make([]Neighbour, 0, len(msg.Neighbors))
		for _, n := range msg.Neighbors {
			list = append(list, Neighbour{Key: n.Pubkey, SNR: n.SNR, HeardSecsAgo: n.HeardSecsAgo})
		}
		store.AddNeighbours(msg.OriginID, at, list)
	case "packets":
		store.Touch(msg.OriginID, at)
		raw, err := hex.DecodeString(msg.Raw)
		if err != nil {
			return
		}
		pt, p, err := splitPacket(raw)
		if err != nil || pt != payloadTypeAdvert {
			return
		}
		a, err := decodeAdvert(p)
		if err != nil {
			st.badAdverts.Add(1)
			slog.Debug("bad advert", "err", err, "topic", topic)
			return
		}
		st.adverts.Add(1)
		store.AddAdvert(a, at)
	}
}

// parseTime reads the observer's timestamp. Times without a zone are taken
// as UTC. Missing or future times become now.
func parseTime(s string, now time.Time) time.Time {
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02T15:04:05.999999999"} {
		if t, err := time.Parse(layout, s); err == nil {
			if t.After(now) {
				return now
			}
			return t
		}
	}
	return now
}
