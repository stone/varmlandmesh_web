package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPublishCreatesBranchWhenMissing(t *testing.T) {
	var calls []string
	var commitBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		if r.Header.Get("Authorization") != "Bearer tok" {
			t.Errorf("missing auth header")
		}
		switch r.Method + " " + r.URL.Path {
		case "POST /repos/o/r/git/trees":
			w.Write([]byte(`{"sha":"tree1"}`))
		case "POST /repos/o/r/git/commits":
			json.NewDecoder(r.Body).Decode(&commitBody)
			w.Write([]byte(`{"sha":"commit1"}`))
		case "PATCH /repos/o/r/git/refs/heads/data":
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte(`{"message":"Reference does not exist"}`))
		case "POST /repos/o/r/git/refs":
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{}`))
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusTeapot)
		}
	}))
	defer srv.Close()

	g := &GitHub{API: srv.URL, Repo: "o/r", Branch: "data", Token: "tok"}
	sha, err := g.Publish(context.Background(), map[string][]byte{"nodes.json": []byte(`{}`)}, "msg")
	if err != nil {
		t.Fatal(err)
	}
	if sha != "commit1" {
		t.Errorf("sha = %s", sha)
	}
	if p, ok := commitBody["parents"].([]any); !ok || len(p) != 0 {
		t.Errorf("commit must have no parents, got %v", commitBody["parents"])
	}
	if len(calls) != 4 {
		t.Errorf("calls = %v", calls)
	}
}

func TestPublishReportsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Bad credentials"}`))
	}))
	defer srv.Close()
	g := &GitHub{API: srv.URL, Repo: "o/r", Branch: "data", Token: "bad"}
	if _, err := g.Publish(context.Background(), map[string][]byte{"nodes.json": nil}, "msg"); err == nil {
		t.Fatal("expected error")
	}
}
