package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/A404coder/launchboard/internal/diagnose"
	"github.com/A404coder/launchboard/internal/launchd"
	"github.com/A404coder/launchboard/internal/server"
	"github.com/A404coder/launchboard/web"
)

// Version is set by goreleaser at build time.
var Version = "dev"

// Config holds CLI flag values.
type Config struct {
	Port   int
	NoOpen bool
}

// Addr returns the listen address string.
func (c Config) Addr() string {
	return fmt.Sprintf("127.0.0.1:%d", c.Port)
}

// newServer creates an http.Server wired with the full router (launchd service,
// diagnose engine, embedded frontend).
func newServer() *http.Server {
	svc := launchd.NewService()
	diag := &diagnose.Engine{}
	router := server.NewRouter(svc, diag, web.FS)
	return &http.Server{Handler: router}
}

func main() {
	var cfg Config
	flag.IntVar(&cfg.Port, "port", 0, "listen port (0 = random available port)")
	flag.BoolVar(&cfg.NoOpen, "no-open", false, "skip auto-opening browser")
	version := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *version {
		fmt.Println("launchboard", Version)
		return
	}

	ln, err := net.Listen("tcp", cfg.Addr())
	if err != nil {
		fmt.Fprintf(os.Stderr, "launchboard: listen: %v\n", err)
		os.Exit(1)
	}

	actualPort := ln.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://127.0.0.1:%d", actualPort)
	fmt.Printf("Launchboard running at %s\n", url)

	srv := newServer()

	// Open browser unless --no-open.
	if !cfg.NoOpen {
		exec.Command("open", url).Start()
	}

	// Start serving in background.
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "launchboard: serve: %v\n", err)
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
		fmt.Fprintf(os.Stderr, "launchboard: shutdown: %v\n", err)
		os.Exit(1)
	}
}
