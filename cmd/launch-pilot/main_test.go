package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"
)

func startTestServer(t *testing.T, port int, noOpen bool) (url string, cancel context.CancelFunc) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())

	cfg := Config{
		Port:         port,
		NoOpen:       noOpen,
		RecentWindow: 10 * time.Minute,
	}

	ln, err := net.Listen("tcp", cfg.Addr())
	if err != nil {
		cancel()
		t.Fatalf("listen: %v", err)
	}

	actualPort := ln.Addr().(*net.TCPAddr).Port
	url = fmt.Sprintf("http://127.0.0.1:%d", actualPort)

	srv := newServer(cfg)

	go func() {
		_ = srv.Serve(ln)
	}()
	go func() {
		<-ctx.Done()
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer shutCancel()
		srv.Shutdown(shutCtx)
	}()

	return url, cancel
}

func TestParseFlags_Defaults(t *testing.T) {
	cfg, versionReq, err := parseFlags([]string{})
	if err != nil {
		t.Fatalf("parseFlags: %v", err)
	}
	if versionReq {
		t.Error("unexpected version request")
	}
	if cfg.RecentWindow != 10*time.Minute {
		t.Errorf("RecentWindow: got %v, want 10m", cfg.RecentWindow)
	}
}

func TestParseFlags_CustomRecentWindow(t *testing.T) {
	cfg, _, err := parseFlags([]string{"--recent-window", "30m"})
	if err != nil {
		t.Fatalf("parseFlags: %v", err)
	}
	if cfg.RecentWindow != 30*time.Minute {
		t.Errorf("RecentWindow: got %v, want 30m", cfg.RecentWindow)
	}
}

func TestParseFlags_RecentWindowTooSmall(t *testing.T) {
	_, _, err := parseFlags([]string{"--recent-window", "30s"})
	if err == nil {
		t.Error("expected error for window below 1m")
	}
	if err != nil && !strings.Contains(err.Error(), "1m and 24h") {
		t.Errorf("error message missing bounds: %v", err)
	}
}

func TestParseFlags_RecentWindowTooLarge(t *testing.T) {
	_, _, err := parseFlags([]string{"--recent-window", "25h"})
	if err == nil {
		t.Error("expected error for window above 24h")
	}
}

func TestParseFlags_RecentWindowBoundaries(t *testing.T) {
	for _, v := range []string{"1m", "24h"} {
		if _, _, err := parseFlags([]string{"--recent-window", v}); err != nil {
			t.Errorf("boundary %s should be accepted: %v", v, err)
		}
	}
}

func TestConfigAddr(t *testing.T) {
	tests := []struct {
		name string
		port int
		want string
	}{
		{"default random port", 0, "127.0.0.1:0"},
		{"explicit port", 8080, "127.0.0.1:8080"},
		{"custom port", 9999, "127.0.0.1:9999"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{Port: tt.port}
			if got := cfg.Addr(); got != tt.want {
				t.Errorf("Addr() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestServerServesHTML(t *testing.T) {
	url, cancel := startTestServer(t, 0, true)
	defer cancel()

	resp, err := http.Get(url + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("GET / status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Launch Pilot") {
		t.Errorf("GET / body does not contain 'Launch Pilot'")
	}
}

func TestServerServesJobsAPI(t *testing.T) {
	url, cancel := startTestServer(t, 0, true)
	defer cancel()

	resp, err := http.Get(url + "/api/jobs")
	if err != nil {
		t.Fatalf("GET /api/jobs: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("GET /api/jobs status = %d, want 200", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}

	if _, ok := result["jobs"]; !ok {
		t.Error("response missing 'jobs' key")
	}
}

func TestServerExplicitPort(t *testing.T) {
	// Find a free port to use as explicit.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	url, cancel := startTestServer(t, port, true)
	defer cancel()

	// Verify it's actually on the expected port.
	expected := fmt.Sprintf("http://127.0.0.1:%d", port)
	if url != expected {
		t.Errorf("url = %q, want %q", url, expected)
	}

	resp, err := http.Get(url + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("GET / status = %d, want 200", resp.StatusCode)
	}
}

func TestStartupBanner_LaunchPilot(t *testing.T) {
	var buf bytes.Buffer
	printBanner(&buf, "http://127.0.0.1:18080")

	line := buf.String()
	matched, err := regexp.MatchString(`^Launch Pilot running at http://127\.0\.0\.1:\d+/?\n$`, line)
	if err != nil {
		t.Fatalf("regex compile: %v", err)
	}
	if !matched {
		t.Errorf("banner mismatch: %q", line)
	}
	if strings.Contains(line, "Launchboard") {
		t.Errorf("banner must not contain 'Launchboard', got %q", line)
	}
}

func TestGracefulShutdown(t *testing.T) {
	url, cancel := startTestServer(t, 0, true)

	// Verify server is up.
	resp, err := http.Get(url + "/")
	if err != nil {
		t.Fatalf("server not up: %v", err)
	}
	resp.Body.Close()

	// Trigger shutdown.
	cancel()

	// Give server time to shut down.
	time.Sleep(100 * time.Millisecond)

	// Server should be down — connection refused.
	client := &http.Client{Timeout: 500 * time.Millisecond}
	_, err = client.Get(url + "/")
	if err == nil {
		t.Error("expected connection error after shutdown, got nil")
	}
}
