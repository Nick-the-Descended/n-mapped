package server

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/nick-the-descended/n-mapped/internal/diff"
	"github.com/nick-the-descended/n-mapped/internal/nmap"
	"github.com/nick-the-descended/n-mapped/internal/store"
)

func (s *Server) routes() error {
	front, err := FrontendFS()
	if err != nil {
		return fmt.Errorf("loading embedded frontend: %w", err)
	}
	s.mux.Handle("/", spaHandler(front))

	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("/api/version", s.handleVersion)
	s.mux.HandleFunc("/api/privilege", s.handlePrivilege)
	s.mux.HandleFunc("/api/catalog", s.handleCatalog)
	s.mux.HandleFunc("/api/scans", s.handleScansCollection)
	s.mux.HandleFunc("/api/scans/", s.handleScanItem)
	s.mux.HandleFunc("/api/history", s.handleHistoryList)
	s.mux.HandleFunc("/api/history/", s.handleHistoryItem)
	s.mux.HandleFunc("/api/favorites", s.handleFavoritesCollection)
	s.mux.HandleFunc("/api/favorites/", s.handleFavoriteItem)
	s.mux.HandleFunc("/api/update", s.handleUpdate)
	s.mux.HandleFunc("/api/diff", s.handleDiff)
	s.mux.HandleFunc("/api/audit", s.handleAudit)
	return nil
}

func (s *Server) handleUpdate(w http.ResponseWriter, _ *http.Request) {
	if s.opts.Update == nil {
		writeJSON(w, http.StatusOK, map[string]any{"enabled": false})
		return
	}
	writeJSON(w, http.StatusOK, s.opts.Update.Status())
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// Connection probably gone; nothing useful to do.
		_ = err
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleVersion(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.opts.NmapInfo)
}

func (s *Server) handlePrivilege(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.opts.Privs)
}

func (s *Server) handleCatalog(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"schema_version": s.opts.Catalog.SchemaVersion,
		"categories":     s.opts.Catalog.Categories,
		"flags":          s.opts.Catalog.Flags,
		"scripts":        s.opts.Catalog.Scripts,
		"profiles":       s.opts.Catalog.Profiles,
	})
}

// POST /api/scans — start a scan.
func (s *Server) handleScansCollection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		writeError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	var req nmap.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	built, err := nmap.Build(req, s.opts.Catalog, s.opts.NmapInfo, s.opts.Privs)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !s.opts.NmapInfo.OK {
		writeError(w, http.StatusFailedDependency, "nmap not detected: "+s.opts.NmapInfo.Error)
		return
	}
	// Detach from the request context: when the POST returns 202 the request
	// context cancels, and we don't want that to kill the running scan.
	scan, err := s.opts.Runner.Start(context.Background(), built)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{
		"id":      scan.ID,
		"argv":    built.Argv,
		"display": built.Display,
		"started": scan.Started,
	})
}

// GET    /api/scans/{id}/events  — SSE event stream
// POST   /api/scans/{id}/stop    — request termination
func (s *Server) handleScanItem(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/api/scans/")
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) != 2 {
		writeError(w, http.StatusNotFound, "expected /api/scans/{id}/{events|stop}")
		return
	}
	id, action := parts[0], parts[1]
	scan, ok := s.opts.Runner.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "no such scan")
		return
	}
	switch action {
	case "events":
		s.streamEvents(w, r, scan)
	case "stop":
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			writeError(w, http.StatusMethodNotAllowed, "POST required")
			return
		}
		scan.Stop()
		writeJSON(w, http.StatusOK, map[string]any{"stopped": true})
	default:
		writeError(w, http.StatusNotFound, "unknown action: "+action)
	}
}

// streamEvents pushes scan events as Server-Sent Events. SSE is one-way and
// stdlib-friendly, which fits scan-output streaming better than WebSockets.
func (s *Server) streamEvents(w http.ResponseWriter, r *http.Request, scan *nmap.Scan) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	events := scan.Subscribe()
	enc := json.NewEncoder(w)
	for {
		select {
		case ev, ok := <-events:
			if !ok {
				return
			}
			if _, err := w.Write([]byte("data: ")); err != nil {
				return
			}
			if err := enc.Encode(ev); err != nil {
				return
			}
			if _, err := w.Write([]byte("\n")); err != nil {
				return
			}
			flusher.Flush()
			if ev.Kind == nmap.EventDone || ev.Kind == nmap.EventError {
				return
			}
		case <-r.Context().Done():
			return
		}
	}
}

// GET /api/history?limit=N — list summaries newest-first.
func (s *Server) handleHistoryList(w http.ResponseWriter, r *http.Request) {
	if s.opts.History == nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	limit := 0
	if q := r.URL.Query().Get("limit"); q != "" {
		if n, err := strconv.Atoi(q); err == nil && n >= 0 {
			limit = n
		}
	}
	list, err := s.opts.History.List(limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// GET    /api/history/{id}     — full record
// DELETE /api/history/{id}     — remove record
// GET    /api/history/{id}/xml — raw nmap XML (text/xml)
func (s *Server) handleHistoryItem(w http.ResponseWriter, r *http.Request) {
	if s.opts.History == nil {
		writeError(w, http.StatusNotFound, "history disabled")
		return
	}
	rest := strings.TrimPrefix(r.URL.Path, "/api/history/")
	if rest == "" {
		writeError(w, http.StatusNotFound, "expected /api/history/{id}")
		return
	}
	parts := strings.SplitN(rest, "/", 2)
	id := parts[0]
	rec, err := s.opts.History.Get(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rec == nil {
		writeError(w, http.StatusNotFound, "no such record")
		return
	}
	if len(parts) == 2 && parts[1] == "xml" {
		if rec.Result == nil || len(rec.Result.RawXML) == 0 {
			writeError(w, http.StatusNotFound, "no XML stored for this record")
			return
		}
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+id+".xml\"")
		_, _ = w.Write(rec.Result.RawXML)
		return
	}
	switch r.Method {
	case http.MethodGet, "":
		writeJSON(w, http.StatusOK, rec)
	case http.MethodDelete:
		if err := s.opts.History.Delete(id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
	case http.MethodPatch:
		var body struct {
			Tags  []string `json:"tags"`
			Notes string   `json:"notes"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		// Cap tag count + length so a malicious or accidental large payload
		// can't bloat the on-disk record.
		if len(body.Tags) > 32 {
			writeError(w, http.StatusBadRequest, "too many tags (max 32)")
			return
		}
		for _, t := range body.Tags {
			if len(t) > 64 {
				writeError(w, http.StatusBadRequest, "tag too long (max 64 chars)")
				return
			}
		}
		if len(body.Notes) > 4096 {
			writeError(w, http.StatusBadRequest, "notes too long (max 4096 chars)")
			return
		}
		if err := s.opts.History.UpdateMeta(id, body.Tags, body.Notes); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		updated, _ := s.opts.History.Get(id)
		writeJSON(w, http.StatusOK, updated)
	default:
		w.Header().Set("Allow", "GET, PATCH, DELETE")
		writeError(w, http.StatusMethodNotAllowed, "GET, PATCH, or DELETE required")
	}
}

// GET  /api/favorites      list
// POST /api/favorites      create or update
func (s *Server) handleFavoritesCollection(w http.ResponseWriter, r *http.Request) {
	if s.opts.Favorites == nil {
		writeError(w, http.StatusNotFound, "favorites disabled")
		return
	}
	switch r.Method {
	case http.MethodGet, "":
		list, err := s.opts.Favorites.List()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, list)
	case http.MethodPost:
		var fav store.Favorite
		if err := json.NewDecoder(r.Body).Decode(&fav); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		saved, err := s.opts.Favorites.Save(fav)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, saved)
	default:
		w.Header().Set("Allow", "GET, POST")
		writeError(w, http.StatusMethodNotAllowed, "GET or POST required")
	}
}

// GET    /api/favorites/{id}    get
// DELETE /api/favorites/{id}    delete
func (s *Server) handleFavoriteItem(w http.ResponseWriter, r *http.Request) {
	if s.opts.Favorites == nil {
		writeError(w, http.StatusNotFound, "favorites disabled")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/favorites/")
	if id == "" {
		writeError(w, http.StatusNotFound, "expected /api/favorites/{id}")
		return
	}
	switch r.Method {
	case http.MethodGet, "":
		fav, err := s.opts.Favorites.Get(id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if fav == nil {
			writeError(w, http.StatusNotFound, "no such favorite")
			return
		}
		writeJSON(w, http.StatusOK, fav)
	case http.MethodDelete:
		if err := s.opts.Favorites.Delete(id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
	default:
		w.Header().Set("Allow", "GET, DELETE")
		writeError(w, http.StatusMethodNotAllowed, "GET or DELETE required")
	}
}

// GET /api/diff?a=<id>&b=<id> — structured diff between two scan records.
func (s *Server) handleDiff(w http.ResponseWriter, r *http.Request) {
	if s.opts.History == nil {
		writeError(w, http.StatusNotFound, "history disabled")
		return
	}
	a := r.URL.Query().Get("a")
	b := r.URL.Query().Get("b")
	if a == "" || b == "" {
		writeError(w, http.StatusBadRequest, "both ?a= and ?b= scan ids are required")
		return
	}
	recA, err := s.opts.History.Get(a)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	recB, err := s.opts.History.Get(b)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if recA == nil || recB == nil {
		writeError(w, http.StatusNotFound, "one or both scan ids not found")
		return
	}
	var runA, runB *nmap.Run
	if recA.Result != nil {
		runA = recA.Result.Run
	}
	if recB.Result != nil {
		runB = recB.Result.Run
	}
	out := diff.Compare(runA, runB)
	writeJSON(w, http.StatusOK, map[string]any{
		"a":          a,
		"b":          b,
		"a_started":  recA.Started,
		"b_started":  recB.Started,
		"a_targets":  recA.Targets,
		"b_targets":  recB.Targets,
		"diff":       out,
	})
}

// GET /api/audit?format=csv|json — flat export of every history record.
// Defaults to JSON. CSV is timestamp-sortable for spreadsheets / SIEM ingest.
func (s *Server) handleAudit(w http.ResponseWriter, r *http.Request) {
	if s.opts.History == nil {
		writeError(w, http.StatusNotFound, "history disabled")
		return
	}
	list, err := s.opts.History.List(0)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}
	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="n-mapped-audit.json"`)
		_ = json.NewEncoder(w).Encode(list)
	case "csv":
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="n-mapped-audit.csv"`)
		writer := csv.NewWriter(w)
		_ = writer.Write([]string{
			"id", "started", "ended", "duration_ms", "exit_code",
			"targets", "hosts_up", "hosts_total", "open_ports",
			"tags", "notes", "command",
		})
		for _, s := range list {
			dur := s.Ended.Sub(s.Started).Milliseconds()
			_ = writer.Write([]string{
				s.ID,
				s.Started.UTC().Format(time.RFC3339),
				s.Ended.UTC().Format(time.RFC3339),
				strconv.FormatInt(dur, 10),
				strconv.Itoa(s.ExitCode),
				strings.Join(s.Targets, " "),
				strconv.Itoa(s.HostsUp),
				strconv.Itoa(s.HostsTotal),
				strconv.Itoa(s.OpenPorts),
				strings.Join(s.Tags, "|"),
				s.Notes,
				s.Display,
			})
		}
		writer.Flush()
	default:
		writeError(w, http.StatusBadRequest, "format must be json or csv")
	}
}

// spaHandler serves the embedded SPA. Unknown non-API paths fall back to
// index.html so the frontend router can handle them.
func spaHandler(front fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(front))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if _, err := fs.Stat(front, path); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				r2 := r.Clone(r.Context())
				r2.URL.Path = "/"
				fileServer.ServeHTTP(w, r2)
				return
			}
		}
		fileServer.ServeHTTP(w, r)
	})
}
