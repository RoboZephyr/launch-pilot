package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/A404coder/launchboard/internal/launchd"
)

// ---------------------------------------------------------------------------
// Mock
// ---------------------------------------------------------------------------

type mockJobService struct {
	jobs    []launchd.Job
	listErr error
}

func (m *mockJobService) ListJobs() ([]launchd.Job, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.jobs, nil
}

func (m *mockJobService) GetJob(label string) (*launchd.Job, error) {
	for i := range m.jobs {
		if m.jobs[i].Label == label {
			return &m.jobs[i], nil
		}
	}
	return nil, fmt.Errorf("job not found: %s", label)
}

func (m *mockJobService) Reload(string) error                          { return nil }
func (m *mockJobService) Start(string) error                           { return nil }
func (m *mockJobService) Stop(string) error                            { return nil }
func (m *mockJobService) ReadLogs(string, int) (*launchd.LogOutput, error) { return nil, nil }

// helper: create a router backed by the mock.
func testRouter(mock *mockJobService) http.Handler {
	return NewRouter(mock, nil, fstest.MapFS{})
}

// ---------------------------------------------------------------------------
// writeJSON / writeError
// ---------------------------------------------------------------------------

func TestWriteJSON_SetsContentType(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, http.StatusOK, map[string]string{"k": "v"})

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}

func TestWriteJSON_SetsStatusCode(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, http.StatusCreated, "ok")

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}
}

func TestWriteJSON_EncodesBody(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, http.StatusOK, map[string]int{"count": 42})

	var body map[string]int
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["count"] != 42 {
		t.Errorf("count = %d, want 42", body["count"])
	}
}

func TestWriteError_ReturnsJSONErrorObject(t *testing.T) {
	w := httptest.NewRecorder()
	writeError(w, http.StatusNotFound, "job not found: com.example.x")

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["error"] != "job not found: com.example.x" {
		t.Errorf("error = %q", body["error"])
	}
}

func TestWriteError_InternalServerError(t *testing.T) {
	w := httptest.NewRecorder()
	writeError(w, http.StatusInternalServerError, "boom")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

// ---------------------------------------------------------------------------
// GET /api/jobs
// ---------------------------------------------------------------------------

func TestListJobs_ReturnsJSONArray(t *testing.T) {
	mock := &mockJobService{
		jobs: []launchd.Job{
			{
				Label:   "com.example.running",
				PID:     584,
				Status:  launchd.StatusRunning,
				Program: "/usr/local/bin/myapp",
				Domain:  "user",
			},
			{
				Label:          "com.example.stopped",
				PID:            0,
				LastExitStatus: 0,
				Status:         launchd.StatusStopped,
				Domain:         "user",
			},
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/jobs", nil)
	testRouter(mock).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q", ct)
	}

	var body struct {
		Jobs      []launchd.Job `json:"jobs"`
		Count     int           `json:"count"`
		Timestamp string        `json:"timestamp"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Count != 2 {
		t.Errorf("count = %d, want 2", body.Count)
	}
	if len(body.Jobs) != 2 {
		t.Errorf("len(jobs) = %d, want 2", len(body.Jobs))
	}
	if body.Timestamp == "" {
		t.Error("timestamp is empty")
	}
	if body.Jobs[0].Label != "com.example.running" {
		t.Errorf("jobs[0].label = %q", body.Jobs[0].Label)
	}
}

func TestListJobs_Empty(t *testing.T) {
	mock := &mockJobService{jobs: []launchd.Job{}}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/jobs", nil)
	testRouter(mock).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var body struct {
		Jobs  []launchd.Job `json:"jobs"`
		Count int           `json:"count"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Count != 0 {
		t.Errorf("count = %d, want 0", body.Count)
	}
	if body.Jobs == nil {
		t.Error("jobs should be [], not null")
	}
}

func TestListJobs_ServiceError(t *testing.T) {
	mock := &mockJobService{listErr: fmt.Errorf("launchctl failed")}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/jobs", nil)
	testRouter(mock).ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["error"] == "" {
		t.Error("error message is empty")
	}
}

// ---------------------------------------------------------------------------
// GET /api/jobs/{label}
// ---------------------------------------------------------------------------

func TestGetJob_Found(t *testing.T) {
	mock := &mockJobService{
		jobs: []launchd.Job{
			{
				Label:   "com.example.myapp",
				PID:     584,
				Status:  launchd.StatusRunning,
				Program: "/usr/local/bin/myapp",
				Domain:  "user",
			},
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/jobs/com.example.myapp", nil)
	testRouter(mock).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var job launchd.Job
	if err := json.NewDecoder(w.Body).Decode(&job); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if job.Label != "com.example.myapp" {
		t.Errorf("label = %q", job.Label)
	}
	if job.PID != 584 {
		t.Errorf("pid = %d, want 584", job.PID)
	}
	if job.Program != "/usr/local/bin/myapp" {
		t.Errorf("program = %q", job.Program)
	}
}

func TestGetJob_NotFound_Returns404(t *testing.T) {
	mock := &mockJobService{
		jobs: []launchd.Job{{Label: "com.example.other"}},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/jobs/com.example.missing", nil)
	testRouter(mock).ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["error"] == "" {
		t.Error("error message is empty")
	}
}

// ---------------------------------------------------------------------------
// Route method enforcement
// ---------------------------------------------------------------------------

func TestRoutes_POSTToGETEndpoint_Returns405(t *testing.T) {
	router := testRouter(&mockJobService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/jobs", nil)
	router.ServeHTTP(w, r)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST /api/jobs: status = %d, want 405", w.Code)
	}
}

func TestRoutes_GETToPostEndpoint_Returns405(t *testing.T) {
	// Use a DELETE method to avoid matching the catch-all GET / file server.
	router := testRouter(&mockJobService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("DELETE", "/api/jobs/com.example.myapp/reload", nil)
	router.ServeHTTP(w, r)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("DELETE /api/jobs/{label}/reload: status = %d, want 405", w.Code)
	}
}
