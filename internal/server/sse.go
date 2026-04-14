package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultSSEInterval = 5 * time.Second

// sseHandlerWithInterval returns an SSE handler that pushes job list events at
// the given interval. The first event is sent immediately upon connection.
func sseHandlerWithInterval(svc JobService, interval time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "SSE not supported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Send first event immediately.
		sendJobsEvent(w, flusher, svc)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case <-ticker.C:
				sendJobsEvent(w, flusher, svc)
			}
		}
	}
}

// sendJobsEvent writes a single SSE event with the current job list.
// On ListJobs error, it sends an "error" event instead.
func sendJobsEvent(w http.ResponseWriter, f http.Flusher, svc JobService) {
	jobs, err := svc.ListJobs()
	if err != nil {
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
		f.Flush()
		return
	}
	data, _ := json.Marshal(jobs)
	fmt.Fprintf(w, "event: jobs\ndata: %s\n\n", data)
	f.Flush()
}
