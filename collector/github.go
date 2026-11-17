package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GitHub publishes files to a branch as a single parentless commit,
// force-updating the branch each time so it never grows history.
type GitHub struct {
	API    string // e.g. https://api.github.com
	Repo   string // owner/name
	Branch string
	Token  string
	Client *http.Client
}

// Publish replaces the branch content with files (path -> content).
func (g *GitHub) Publish(ctx context.Context, files map[string][]byte, message string) (string, error) {
	type entry struct {
		Path    string `json:"path"`
		Mode    string `json:"mode"`
		Type    string `json:"type"`
		Content string `json:"content"`
	}
	var tree []entry
	for p, c := range files {
		tree = append(tree, entry{Path: p, Mode: "100644", Type: "blob", Content: string(c)})
	}

	var t struct{ SHA string }
	if _, err := g.call(ctx, "POST", "/git/trees", map[string]any{"tree": tree}, &t); err != nil {
		return "", fmt.Errorf("create tree: %w", err)
	}
	var c struct{ SHA string }
	if _, err := g.call(ctx, "POST", "/git/commits", map[string]any{"message": message, "tree": t.SHA, "parents": []string{}}, &c); err != nil {
		return "", fmt.Errorf("create commit: %w", err)
	}

	status, err := g.call(ctx, "PATCH", "/git/refs/heads/"+g.Branch, map[string]any{"sha": c.SHA, "force": true}, nil)
	if status == http.StatusUnprocessableEntity || status == http.StatusNotFound {
		// Branch does not exist yet.
		_, err = g.call(ctx, "POST", "/git/refs", map[string]any{"ref": "refs/heads/" + g.Branch, "sha": c.SHA}, nil)
	}
	if err != nil {
		return "", fmt.Errorf("update ref: %w", err)
	}
	return c.SHA, nil
}

func (g *GitHub) call(ctx context.Context, method, path string, body, out any) (int, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequestWithContext(ctx, method, g.API+"/repos/"+g.Repo+path, bytes.NewReader(b))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+g.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "varmlandmesh-collector")

	client := g.Client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode/100 != 2 {
		return res.StatusCode, fmt.Errorf("%s %s: %s: %s", method, path, res.Status, bytes.TrimSpace(data))
	}
	if out != nil {
		if err := json.Unmarshal(data, out); err != nil {
			return res.StatusCode, err
		}
	}
	return res.StatusCode, nil
}
