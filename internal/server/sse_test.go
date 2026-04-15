package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/A404coder/launch-pilot/internal/diagnose"
	"github.com/A404coder/launch-pilot/internal/launchd"
)

// ---------------------------------------------------------------------------
// SSE test helpers
// ---------------------------------------------------------------------------

// sseEvent represents a parsed Server-Sent Event.
type sseEvent struct {
	Event string
	Data  string
}

// readSSEEvents reads events from the response body until the context is cancelled
// or limit events are read, whichever comes first.
func readSSEEvents(t *testing.T, body *strings.Reader, limit int) []sseEvent {
	t.Helper()
	var events []sseEvent
	scanner := bufio.NewScanner(body)
	var current sseEvent
	for scanner.Scan() && len(events) < limit {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "event: "):
			current.Event = strings.TrimPrefix(line, "event: ")
		case strings.HasPrefix(line, "data: "):
			current.Data = strings.TrimPrefix(line, "data: ")
		case line == "":
			if current.Event != "" || current.Data != "" {
				events = append(events, current)
				current = sseEvent{}
			}
		}
	}
	return events
}

// ---------------------------------------------------------------------------
// Content-Type and Cache-Control headers
// ---------------------------------------------------------------------------

func TestSSE_ContentType_IsEventStream(t *testing.T) {
	mock := &mockJobService{jobs: []launchd.Job{}}
	router := NewRouter(mock, &diagnose.Engine{}, fstest.MapFS{})

	// Use a context that cancels quickly so the handler exits after the first event.
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(ctx, "GET", "/api/events", nil)
	router.ServeHTTP(w, r)

	ct := w.Header().Get("Content-Type")
	if ct != "text/event-stream" {
		t.Errorf("Content-Type = %q, want text/event-stream", ct)
	}
}

func TestSSE_CacheControl_IsNoCache(t *testing.T) {
	mock := &mockJobService{jobs: []launchd.Job{}}
	router := NewRouter(mock, &diagnose.Engine{}, fstest.MapFS{})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(ctx, "GET", "/api/events", nil)
	router.ServeHTTP(w, r)

	cc := w.Header().Get("Cache-Control")
	if cc != "no-cache" {
		t.Errorf("Cache-Control = %q, want no-cache", cc)
	}
}

func TestSSE_ConnectionHeader_IsKeepAlive(t *testing.T) {
	mock := &mockJobService{jobs: []launchd.Job{}}
	router := NewRouter(mock, &diagnose.Engine{}, fstest.MapFS{})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(ctx, "GET", "/api/events", nil)
	router.ServeHTTP(w, r)

	conn := w.Header().Get("Connection")
	if conn != "keep-alive" {
		t.Errorf("Connection = %q, want keep-alive", conn)
	}
}

// ---------------------------------------------------------------------------
// First event sent immediately
// ---------------------------------------------------------------------------

func TestSSE_FirstEventSentImmediately(t *testing.T) {
	mock := &mockJobService{
		jobs: []launchd.Job{
			{Label: "com.example.app", PID: 123, Status: launchd.StatusRunning},
		},
	}
	router := NewRouter(mock, &diagnose.Engine{}, fstest.MapFS{})

	// Cancel quickly — we only need the first event.
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(ctx, "GET", "/api/events", nil)
	router.ServeHTTP(w, r)

	events := readSSEEvents(t, strings.NewReader(w.Body.String()), 1)
	if len(events) == 0 {
		t.Fatal("expected at least one event, got none")
	}

	first := events[0]
	if first.Event != "jobs" {
		t.Errorf("first event type = %q, want jobs", first.Event)
	}
}

// ---------------------------------------------------------------------------
// Event data is valid JSON array of jobs
// ---------------------------------------------------------------------------

func TestSSE_EventDataIsValidJobsJSON(t *testing.T) {
	mock := &mockJobService{
		jobs: []launchd.Job{
			{Label: "com.example.running", PID: 584, Status: launchd.StatusRunning},
			{Label: "com.example.stopped", PID: 0, Status: launchd.StatusStopped},
		},
	}
	router := NewRouter(mock, &diagnose.Engine{}, fstest.MapFS{})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(ctx, "GET", "/api/events", nil)
	router.ServeHTTP(w, r)

	events := readSSEEvents(t, strings.NewReader(w.Body.String()), 1)
	if len(events) == 0 {
		t.Fatal("expected at least one event")
	}

	var jobs []launchd.Job
	if err := json.Unmarshal([]byte(events[0].Data), &jobs); err != nil {
		t.Fatalf("failed to parse event data as JSON: %v\ndata: %s", err, events[0].Data)
	}
	if len(jobs) != 2 {
		t.Errorf("len(jobs) = %d, want 2", len(jobs))
	}
	if jobs[0].Label != "com.example.running" {
		t.Errorf("jobs[0].label = %q, want com.example.running", jobs[0].Label)
	}
}

// ---------------------------------------------------------------------------
// Handler exits when context is cancelled
// ---------------------------------------------------------------------------

func TestSSE_ExitsOnContextCancel(t *testing.T) {
	mock := &mockJobService{jobs: []launchd.Job{}}
	router := NewRouter(mock, &diagnose.Engine{}, fstest.MapFS{})

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(ctx, "GET", "/api/events", nil)

	go func() {
		router.ServeHTTP(w, r)
		close(done)
	}()

	// Cancel the context after a short delay.
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Handler exited — success.
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not exit after context cancellation")
	}
}

// ---------------------------------------------------------------------------
// Error event on ListJobs failure
// ---------------------------------------------------------------------------

func TestSSE_ListJobsError_SendsErrorEvent(t *testing.T) {
	mock := &mockJobService{
		listErr: fmt.Errorf("launchctl failed"),
	}
	router := NewRouter(mock, &diagnose.Engine{}, fstest.MapFS{})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(ctx, "GET", "/api/events", nil)
	router.ServeHTTP(w, r)

	events := readSSEEvents(t, strings.NewReader(w.Body.String()), 1)
	if len(events) == 0 {
		t.Fatal("expected at least one event")
	}

	if events[0].Event != "error" {
		t.Errorf("event type = %q, want error", events[0].Event)
	}
	if !strings.Contains(events[0].Data, "launchctl failed") {
		t.Errorf("error data = %q, should contain 'launchctl failed'", events[0].Data)
	}
}

// ---------------------------------------------------------------------------
// SSE event format: "event: jobs\ndata: ...\n\n"
// ---------------------------------------------------------------------------

func TestSSE_EventFormat_HasEventAndDataLines(t *testing.T) {
	mock := &mockJobService{jobs: []launchd.Job{}}
	router := NewRouter(mock, &diagnose.Engine{}, fstest.MapFS{})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(ctx, "GET", "/api/events", nil)
	router.ServeHTTP(w, r)

	body := w.Body.String()

	// Must contain "event: jobs" line.
	if !strings.Contains(body, "event: jobs") {
		t.Errorf("response body missing 'event: jobs' line:\n%s", body)
	}

	// Must contain "data: " line.
	if !strings.Contains(body, "data: ") {
		t.Errorf("response body missing 'data: ' line:\n%s", body)
	}

	// Events are terminated by a blank line (\n\n).
	if !strings.Contains(body, "\n\n") {
		t.Errorf("response body missing event terminator (blank line):\n%s", body)
	}
}

// ---------------------------------------------------------------------------
// Multiple events over time (ticker fires)
// ---------------------------------------------------------------------------

func TestSSE_MultipleTickEvents(t *testing.T) {
	mock := &mockJobService{
		jobs: []launchd.Job{
			{Label: "com.example.app", PID: 1, Status: launchd.StatusRunning},
		},
	}

	// Use a short interval for testing.
	h := sseHandlerWithInterval(mock, 50*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(ctx, "GET", "/api/events", nil)
	h.ServeHTTP(w, r)

	events := readSSEEvents(t, strings.NewReader(w.Body.String()), 10)
	// We expect at least 2 events: 1 immediate + 1+ from ticker in 200ms with 50ms interval.
	if len(events) < 2 {
		t.Errorf("expected >= 2 events, got %d", len(events))
	}
}

// ---------------------------------------------------------------------------
// Empty jobs list produces valid JSON
// ---------------------------------------------------------------------------

func TestSSE_EmptyJobs_ProducesEmptyArray(t *testing.T) {
	mock := &mockJobService{jobs: []launchd.Job{}}
	router := NewRouter(mock, &diagnose.Engine{}, fstest.MapFS{})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(ctx, "GET", "/api/events", nil)
	router.ServeHTTP(w, r)

	events := readSSEEvents(t, strings.NewReader(w.Body.String()), 1)
	if len(events) == 0 {
		t.Fatal("expected at least one event")
	}

	var jobs []launchd.Job
	if err := json.Unmarshal([]byte(events[0].Data), &jobs); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if jobs == nil {
		t.Error("jobs should be [], not null")
	}
	if len(jobs) != 0 {
		t.Errorf("len(jobs) = %d, want 0", len(jobs))
	}
}
