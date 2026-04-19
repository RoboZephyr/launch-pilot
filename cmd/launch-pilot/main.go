package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/A404coder/launch-pilot/internal/diagnose"
	"github.com/A404coder/launch-pilot/internal/launchd"
	"github.com/A404coder/launch-pilot/internal/server"
	"github.com/A404coder/launch-pilot/web"
)

// Version is set by goreleaser at build time.
var Version = "dev"

// Config holds CLI flag values.
type Config struct {
	Port         int
	NoOpen       bool
	RecentWindow time.Duration
}

// Addr returns the listen address string.
func (c Config) Addr() string {
	return fmt.Sprintf("127.0.0.1:%d", c.Port)
}

// Bounds for --recent-window per spec (1m to 24h).
const (
	minRecentWindow = time.Minute
	maxRecentWindow = 24 * time.Hour
)

// parseFlags parses command-line arguments into a Config and validates them.
// Returns a non-nil versionRequested when --version was passed.
func parseFlags(args []string) (cfg Config, versionRequested bool, err error) {
	fs := flag.NewFlagSet("launch-pilot", flag.ContinueOnError)
	fs.IntVar(&cfg.Port, "port", 0, "listen port (0 = random available port)")
	fs.BoolVar(&cfg.NoOpen, "no-open", false, "skip auto-opening browser")
	fs.DurationVar(&cfg.RecentWindow, "recent-window", launchd.DefaultRecentWindow,
		"how long a just-finished job is shown as 'completed' (1m–24h)")
	versionFlag := fs.Bool("version", false, "print version and exit")
	if err := fs.Parse(args); err != nil {
		return cfg, false, err
	}
	if *versionFlag {
		return cfg, true, nil
	}
	if cfg.RecentWindow < minRecentWindow || cfg.RecentWindow > maxRecentWindow {
		return cfg, false, fmt.Errorf("--recent-window must be between 1m and 24h, got %v", cfg.RecentWindow)
	}
	return cfg, false, nil
}

// printBanner writes the startup banner announcing the running URL.
func printBanner(w io.Writer, url string) {
	fmt.Fprintf(w, "Launch Pilot running at %s\n", url)
}

// newServer creates an http.Server wired with the full router (launchd service,
// diagnose engine, embedded frontend).
func newServer(cfg Config) *http.Server {
	svc := launchd.NewServiceWithWindow(cfg.RecentWindow)
	diag := &diagnose.Engine{}
	router := server.NewRouter(svc, diag, web.FS)
	return &http.Server{Handler: router}
}

func main() {
	cfg, versionRequested, err := parseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "launch-pilot:", err)
		os.Exit(1)
	}
	if versionRequested {
		fmt.Println("launch-pilot", Version)
		return
	}

	ln, err := net.Listen("tcp", cfg.Addr())
	if err != nil {
		fmt.Fprintf(os.Stderr, "launch-pilot: listen: %v\n", err)
		os.Exit(1)
	}

	actualPort := ln.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://127.0.0.1:%d", actualPort)
	printBanner(os.Stdout, url)

	srv := newServer(cfg)

	// Open browser unless --no-open.
	if !cfg.NoOpen {
		exec.Command("open", url).Start()
	}

	// Start serving in background.
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "launch-pilot: serve: %v\n", err)
		}
	}()

	// Wait for interrupt signal.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	stop() // restore default signal behavior

	fmt.Println("\nShutting down...")
	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutCtx); err != nil {
		fmt.Fprintf(os.Stderr, "launch-pilot: shutdown: %v\n", err)
		os.Exit(1)
	}
}
