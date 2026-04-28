// Package server wires the HTTP API and the embedded SPA together.
package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/nick-the-descended/n-mapped/internal/auth"
	"github.com/nick-the-descended/n-mapped/internal/catalog"
	"github.com/nick-the-descended/n-mapped/internal/nmap"
	"github.com/nick-the-descended/n-mapped/internal/store"
)

// Options bundles everything the server needs at construction time.
type Options struct {
	Bind     string // e.g. "127.0.0.1:8765"
	NmapPath string // override for the nmap binary; "" means PATH lookup
	Catalog  *catalog.Catalog
	NmapInfo nmap.Info
	Privs    auth.State
	Runner   *nmap.Runner
	History  *store.History
}

// Server is an http.Handler plus the lifecycle hooks for graceful shutdown.
type Server struct {
	opts Options
	mux  *http.ServeMux
	srv  *http.Server
}

// New constructs a Server. ListenAndServe must still be called.
func New(opts Options) (*Server, error) {
	if opts.Bind == "" {
		opts.Bind = "127.0.0.1:8765"
	}
	if opts.Catalog == nil {
		return nil, errors.New("server: catalog is required")
	}
	if opts.Runner == nil {
		opts.Runner = nmap.NewRunner()
	}
	if opts.History != nil {
		hist := opts.History
		opts.Runner.SetHook(func(res nmap.Result) {
			rec := store.RecordFromResult(res)
			if err := hist.Save(rec); err != nil {
				log.Printf("history: save scan %s failed: %v", res.ID, err)
			}
		})
	}
	s := &Server{opts: opts, mux: http.NewServeMux()}
	if err := s.routes(); err != nil {
		return nil, err
	}
	s.srv = &http.Server{
		Addr:              opts.Bind,
		Handler:           s.mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	return s, nil
}

// URL returns the human-readable URL the user should open.
func (s *Server) URL() string {
	return fmt.Sprintf("http://%s", s.opts.Bind)
}

// ListenAndServe blocks until the underlying http.Server returns.
func (s *Server) ListenAndServe() error {
	log.Printf("n-mapped listening on %s", s.URL())
	return s.srv.ListenAndServe()
}

// Shutdown gracefully closes the listener.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
