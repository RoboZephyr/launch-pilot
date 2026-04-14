package server

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/A404coder/launchboard/internal/diagnose"
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

// labelRe validates launchd job labels: alphanumeric, dots, hyphens, underscores.
var labelRe = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

func validLabel(label string) bool {
	return labelRe.MatchString(label)
}

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
		job, err := svc.GetJob(label)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
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
		if !validLabel(label) {
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
			writeJSON(w, http.StatusInternalServerError, map[string]any{
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
		if !validLabel(label) {
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
			lines = n
		}

		logs, err := svc.ReadLogs(label, lines)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
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
		if !validLabel(label) {
			writeError(w, http.StatusBadRequest, "invalid label format")
			return
		}

		job, err := svc.GetJob(label)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
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
