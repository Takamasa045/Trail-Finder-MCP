package httpx

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestRetriesOn503ThenSucceeds(t *testing.T) {
	var n atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if n.Add(1) < 3 {
			http.Error(w, "busy", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer ts.Close()

	c := &Client{HTTP: ts.Client(), Retries: 2, Backoff: time.Millisecond}

	var out map[string]any
	if err := c.GetJSON(context.Background(), ts.URL, &out); err != nil {
		t.Fatalf("GetJSON: %v", err)
	}
	if out["ok"] != true {
		t.Fatalf("out=%v", out)
	}
	if n.Load() != 3 {
		t.Fatalf("attempts=%d", n.Load())
	}
}

func TestDoesNotRetryOn400(t *testing.T) {
	var n atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n.Add(1)
		http.Error(w, "nope", http.StatusBadRequest)
	}))
	defer ts.Close()

	c := &Client{HTTP: ts.Client(), Retries: 2, Backoff: time.Millisecond}

	err := c.GetJSON(context.Background(), ts.URL, &map[string]any{})
	if err == nil {
		t.Fatal("expected error")
	}
	if n.Load() != 1 {
		t.Fatalf("attempts=%d", n.Load())
	}
}

func TestPostFormRewindsBody(t *testing.T) {
	var n atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		if string(b) != "data=hello" {
			http.Error(w, "bad body "+string(b), http.StatusBadRequest)
			return
		}
		if n.Add(1) < 2 {
			http.Error(w, "busy", http.StatusBadGateway)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"got": "hello"})
	}))
	defer ts.Close()

	c := &Client{HTTP: ts.Client(), Retries: 2, Backoff: time.Millisecond}

	var out map[string]string
	if err := c.PostFormJSON(context.Background(), ts.URL, []byte("data=hello"), &out); err != nil {
		t.Fatalf("PostFormJSON: %v", err)
	}
	if out["got"] != "hello" {
		t.Fatalf("out=%v", out)
	}
}

func TestGetJSONNilDestAndCancel(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer ts.Close()
	c := &Client{HTTP: ts.Client(), Retries: 0}
	if err := c.GetJSON(context.Background(), ts.URL, nil); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "busy", http.StatusBadGateway)
	}))
	defer slow.Close()
	c2 := &Client{HTTP: slow.Client(), Retries: 2, Backoff: 50 * time.Millisecond}
	if err := c2.GetJSON(ctx, slow.URL, &map[string]any{}); err == nil {
		t.Fatal("expected cancel or status error")
	}
}

func TestSetsUserAgent(t *testing.T) {
	t.Setenv("TRAILFINDER_USER_AGENT", "trail-finder-test/0")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "trail-finder-test/0" {
			http.Error(w, "ua="+r.Header.Get("User-Agent"), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	c := &Client{HTTP: ts.Client(), Retries: 0}
	if err := c.GetJSON(context.Background(), ts.URL, &map[string]any{}); err != nil {
		t.Fatal(err)
	}
}
