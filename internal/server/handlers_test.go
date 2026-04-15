package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/A404coder/launch-pilot/internal/diagnose"
	"github.com/A404coder/launch-pilot/internal/launchd"
)

// ---------------------------------------------------------------------------
// Mock
// ---------------------------------------------------------------------------

type mockJobService struct {
	jobs      []launchd.Job
	listErr   error
	reloadErr error
	startErr  error
	stopErr   error
	logs      *launchd.LogOutput
	logsErr   error

	// Track ReadLogs calls to verify default lines parameter.
	lastReadLogsLines int
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
	return nil, fmt.Errorf("%w: %s", launchd.ErrNotFound, label)
}

func (m *mockJobService) Reload(string) error { return m.reloadErr }
func (m *mockJobService) Start(string) error  { return m.startErr }
func (m *mockJobService) Stop(string) error   { return m.stopErr }
func (m *mockJobService) ReadLogs(_ string, lines int) (*launchd.LogOutput, error) {
	m.lastReadLogsLines = lines
	if m.logsErr != nil {
		return nil, m.logsErr
	}
	return m.logs, nil
}

// helper: create a router backed by the mock (no diagnose engine).
func testRouter(mock *mockJobService) http.Handler {
	return NewRouter(mock, nil, fstest.MapFS{})
}

// helper: create a router with a diagnose engine.
func testRouterWithDiag(mock *mockJobService) http.Handler {
	return NewRouter(mock, &diagnose.Engine{}, fstest.MapFS{})
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

// ---------------------------------------------------------------------------
// POST /api/jobs/{label}/reload — label validation
// ---------------------------------------------------------------------------

func TestActionReload_InvalidLabel_Returns400(t *testing.T) {
	router := testRouter(&mockJobService{})

	// Labels with characters outside [a-zA-Z0-9._-] — must also be valid URL path segments.
	for _, label := range []string{"semi;colon", "bad!label", "@root", "has+plus"} {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/jobs/"+label+"/reload", nil)
		router.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("POST /api/jobs/%s/reload: status = %d, want 400", label, w.Code)
		}
	}
}

func TestActionStart_InvalidLabel_Returns400(t *testing.T) {
	router := testRouter(&mockJobService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/jobs/bad!label/start", nil)
	router.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestActionStop_InvalidLabel_Returns400(t *testing.T) {
	router := testRouter(&mockJobService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/jobs/bad!label/stop", nil)
	router.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// ---------------------------------------------------------------------------
// POST /api/jobs/{label}/start — success
// ---------------------------------------------------------------------------

func TestActionStart_Success_Returns200(t *testing.T) {
	router := testRouter(&mockJobService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/jobs/com.example.myapp/start", nil)
	router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["ok"] != true {
		t.Errorf("ok = %v, want true", body["ok"])
	}
	if body["action"] != "start" {
		t.Errorf("action = %v, want start", body["action"])
	}
	if body["label"] != "com.example.myapp" {
		t.Errorf("label = %v, want com.example.myapp", body["label"])
	}
}

// ---------------------------------------------------------------------------
// POST /api/jobs/{label}/stop — success
// ---------------------------------------------------------------------------

func TestActionStop_Success_Returns200(t *testing.T) {
	router := testRouter(&mockJobService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/jobs/com.example.myapp/stop", nil)
	router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["ok"] != true {
		t.Errorf("ok = %v, want true", body["ok"])
	}
	if body["action"] != "stop" {
		t.Errorf("action = %v, want stop", body["action"])
	}
}

// ---------------------------------------------------------------------------
// POST /api/jobs/{label}/reload — service error
// ---------------------------------------------------------------------------

func TestActionReload_ServiceError_Returns500(t *testing.T) {
	mock := &mockJobService{reloadErr: fmt.Errorf("bootstrap failed: path not found")}
	router := testRouter(mock)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/jobs/com.example.myapp/reload", nil)
	router.ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["ok"] != false {
		t.Errorf("ok = %v, want false", body["ok"])
	}
	if body["error"] == nil || body["error"] == "" {
		t.Error("error message is empty")
	}
}

// ---------------------------------------------------------------------------
// GET /api/jobs/{label}/logs
// ---------------------------------------------------------------------------

func TestGetLogs_DefaultLines200(t *testing.T) {
	stdout := "line1\nline2"
	mock := &mockJobService{
		logs: &launchd.LogOutput{
			Label:           "com.example.myapp",
			Stdout:          &stdout,
			StdoutPath:      "/tmp/out.log",
			StdoutAvailable: true,
		},
	}
	router := testRouter(mock)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/jobs/com.example.myapp/logs", nil)
	router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if mock.lastReadLogsLines != 200 {
		t.Errorf("ReadLogs lines = %d, want 200", mock.lastReadLogsLines)
	}
}

func TestGetLogs_CustomLines(t *testing.T) {
	stdout := "line1"
	mock := &mockJobService{
		logs: &launchd.LogOutput{
			Label:           "com.example.myapp",
			Stdout:          &stdout,
			StdoutAvailable: true,
		},
	}
	router := testRouter(mock)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/jobs/com.example.myapp/logs?lines=50", nil)
	router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if mock.lastReadLogsLines != 50 {
		t.Errorf("ReadLogs lines = %d, want 50", mock.lastReadLogsLines)
	}
}

func TestGetLogs_InvalidLinesParam_Returns400(t *testing.T) {
	mock := &mockJobService{}
	router := testRouter(mock)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/jobs/com.example.myapp/logs?lines=abc", nil)
	router.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestGetLogs_NotFound_Returns404(t *testing.T) {
	mock := &mockJobService{
		logsErr: fmt.Errorf("%w: com.example.missing", launchd.ErrNotFound),
	}
	router := testRouter(mock)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/jobs/com.example.missing/logs", nil)
	router.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestGetLogs_ReturnsLogOutputJSON(t *testing.T) {
	stdout := "hello stdout"
	stderr := "hello stderr"
	mock := &mockJobService{
		logs: &launchd.LogOutput{
			Label:           "com.example.myapp",
			Stdout:          &stdout,
			Stderr:          &stderr,
			StdoutPath:      "/tmp/out.log",
			StderrPath:      "/tmp/err.log",
			StdoutAvailable: true,
			StderrAvailable: true,
		},
	}
	router := testRouter(mock)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/jobs/com.example.myapp/logs", nil)
	router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var body launchd.LogOutput
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Label != "com.example.myapp" {
		t.Errorf("label = %q", body.Label)
	}
	if body.Stdout == nil || *body.Stdout != "hello stdout" {
		t.Errorf("stdout = %v", body.Stdout)
	}
	if body.Stderr == nil || *body.Stderr != "hello stderr" {
		t.Errorf("stderr = %v", body.Stderr)
	}
	if !body.StdoutAvailable {
		t.Error("stdoutAvailable = false, want true")
	}
}

func TestGetLogs_InvalidLabel_Returns400(t *testing.T) {
	router := testRouter(&mockJobService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/jobs/bad!label/logs", nil)
	router.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// ---------------------------------------------------------------------------
// GET /api/jobs/{label}/diagnose
// ---------------------------------------------------------------------------

func TestDiagnose_ReturnsReportWith6Checks(t *testing.T) {
	mock := &mockJobService{
		jobs: []launchd.Job{
			{
				Label:   "com.example.myapp",
				PID:     0,
				Status:  launchd.StatusError,
				Program: "/usr/bin/true", // exists on macOS
				Domain:  "user",
			},
		},
	}
	router := testRouterWithDiag(mock)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/jobs/com.example.myapp/diagnose", nil)
	router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var report diagnose.DiagnoseReport
	if err := json.NewDecoder(w.Body).Decode(&report); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if report.Label != "com.example.myapp" {
		t.Errorf("label = %q, want com.example.myapp", report.Label)
	}
	if len(report.Checks) != 6 {
		t.Fatalf("len(checks) = %d, want 6", len(report.Checks))
	}

	// Verify all expected check IDs are present.
	wantIDs := map[string]bool{
		"exit-code":          false,
		"program-exists":     false,
		"program-executable": false,
		"plist-owner":        false,
		"plist-perms":        false,
		"log-path-exists":    false,
	}
	for _, c := range report.Checks {
		if _, ok := wantIDs[c.ID]; ok {
			wantIDs[c.ID] = true
		} else {
			t.Errorf("unexpected check ID: %q", c.ID)
		}
	}
	for id, found := range wantIDs {
		if !found {
			t.Errorf("missing check ID: %q", id)
		}
	}
}

func TestDiagnose_NotFound_Returns404(t *testing.T) {
	mock := &mockJobService{jobs: []launchd.Job{}}
	router := testRouterWithDiag(mock)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/jobs/com.example.missing/diagnose", nil)
	router.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestDiagnose_InvalidLabel_Returns400(t *testing.T) {
	router := testRouterWithDiag(&mockJobService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/jobs/bad!label/diagnose", nil)
	router.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}
