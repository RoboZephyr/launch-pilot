package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/A404coder/launchboard/internal/diagnose"
	"github.com/A404coder/launchboard/internal/launchd"
)

// writeJSON marshals v as JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError writes a JSON error response: {"error": "<message>"}.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

const maxLogLines = 10000

func listJobsHandler(svc JobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobs, err := svc.ListJobs()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"jobs":      jobs,
			"count":     len(jobs),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}
}

func getJobHandler(svc JobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		label := r.PathValue("label")
		if !launchd.ValidLabel(label) {
			writeError(w, http.StatusBadRequest, "invalid label format")
			return
		}
		job, err := svc.GetJob(label)
		if err != nil {
			if errors.Is(err, launchd.ErrNotFound) {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, job)
	}
}

func actionHandler(svc JobService, action string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		label := r.PathValue("label")
		if !launchd.ValidLabel(label) {
			writeError(w, http.StatusBadRequest, "invalid label format")
			return
		}

		var err error
		switch action {
		case "reload":
			err = svc.Reload(label)
		case "start":
			err = svc.Start(label)
		case "stop":
			err = svc.Stop(label)
		}

		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, launchd.ErrNotFound) || errors.Is(err, launchd.ErrInvalidLabel) {
				status = http.StatusBadRequest
			}
			writeJSON(w, status, map[string]any{
				"ok":     false,
				"label":  label,
				"action": action,
				"error":  err.Error(),
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":     true,
			"label":  label,
			"action": action,
		})
	}
}

func getLogsHandler(svc JobService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		label := r.PathValue("label")
		if !launchd.ValidLabel(label) {
			writeError(w, http.StatusBadRequest, "invalid label format")
			return
		}

		lines := 200
		if q := r.URL.Query().Get("lines"); q != "" {
			n, err := strconv.Atoi(q)
			if err != nil || n < 1 {
				writeError(w, http.StatusBadRequest, "invalid lines parameter")
				return
			}
			if n > maxLogLines {
				n = maxLogLines
			}
			lines = n
		}

		logs, err := svc.ReadLogs(label, lines)
		if err != nil {
			if errors.Is(err, launchd.ErrNotFound) {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, logs)
	}
}

func diagnoseHandler(svc JobService, diag *diagnose.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		label := r.PathValue("label")
		if !launchd.ValidLabel(label) {
			writeError(w, http.StatusBadRequest, "invalid label format")
			return
		}

		job, err := svc.GetJob(label)
		if err != nil {
			if errors.Is(err, launchd.ErrNotFound) {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		report := diag.Diagnose(job)
		writeJSON(w, http.StatusOK, report)
	}
}

// sseHandler returns an SSE handler using the default 5-second push interval.
func sseHandler(svc JobService) http.HandlerFunc {
	return sseHandlerWithInterval(svc, defaultSSEInterval)
}
