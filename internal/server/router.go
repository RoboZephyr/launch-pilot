package server

import (
	"io/fs"
	"net/http"

	"github.com/A404coder/launchboard/internal/diagnose"
	"github.com/A404coder/launchboard/internal/launchd"
)

// JobService defines the launchd operations used by API handlers.
type JobService interface {
	ListJobs() ([]launchd.Job, error)
	GetJob(label string) (*launchd.Job, error)
	Reload(label string) error
	Start(label string) error
	Stop(label string) error
	ReadLogs(label string, lines int) (*launchd.LogOutput, error)
}

// NewRouter creates the HTTP handler with all API routes and static file serving.
func NewRouter(svc JobService, diag *diagnose.Engine, webFS fs.FS) http.Handler {
	mux := http.NewServeMux()

	// Jobs API
	mux.HandleFunc("GET /api/jobs", listJobsHandler(svc))
	mux.HandleFunc("GET /api/jobs/{label}", getJobHandler(svc))

	// Job actions
	mux.HandleFunc("POST /api/jobs/{label}/reload", actionHandler(svc, "reload"))
	mux.HandleFunc("POST /api/jobs/{label}/start", actionHandler(svc, "start"))
	mux.HandleFunc("POST /api/jobs/{label}/stop", actionHandler(svc, "stop"))

	// Logs + diagnostics
	mux.HandleFunc("GET /api/jobs/{label}/logs", getLogsHandler(svc))
	mux.HandleFunc("GET /api/jobs/{label}/diagnose", diagnoseHandler(svc, diag))

	// SSE real-time push
	mux.HandleFunc("GET /api/events", sseHandler(svc))

	// Static files — embedded frontend
	mux.Handle("GET /", http.FileServerFS(webFS))

	return mux
}
