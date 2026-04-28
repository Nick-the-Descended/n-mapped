// Command n-mapped is a beginner-friendly local GUI for Nmap.
//
// Launches an HTTP server on 127.0.0.1, embeds the Svelte frontend, and
// drives the system `nmap` binary as a subprocess. Privilege is decided at
// launch — pass --privileged or run via sudo to enable raw-socket scans.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nick-the-descended/n-mapped/internal/auth"
	"github.com/nick-the-descended/n-mapped/internal/catalog"
	"github.com/nick-the-descended/n-mapped/internal/config"
	"github.com/nick-the-descended/n-mapped/internal/nmap"
	"github.com/nick-the-descended/n-mapped/internal/server"
	"github.com/nick-the-descended/n-mapped/internal/store"
)

// Version is set at build time via -ldflags "-X main.Version=...".
var Version = "dev"

func main() {
	bind := flag.String("bind", "127.0.0.1:8765", "host:port to bind the HTTP server")
	nmapPath := flag.String("nmap-path", "", "override the nmap binary path (default: PATH lookup)")
	privileged := flag.Bool("privileged", false, "advisory: indicate elevated privileges are intended (does not itself elevate; run via sudo)")
	noBrowser := flag.Bool("no-browser", false, "do not auto-open a browser tab on startup")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("n-mapped", Version)
		return
	}

	if err := run(*bind, *nmapPath, *privileged, *noBrowser); err != nil {
		log.Fatalf("n-mapped: %v", err)
	}
}

func run(bind, nmapPath string, privileged, noBrowser bool) error {
	dataDir, err := config.DataDir()
	if err != nil {
		return fmt.Errorf("resolving data dir: %w", err)
	}
	log.Printf("data dir: %s", dataDir)

	cat, err := catalog.Load()
	if err != nil {
		return fmt.Errorf("loading catalog: %w", err)
	}
	log.Printf("catalog: schema %s, %d categories, %d flags", cat.SchemaVersion, len(cat.Categories), len(cat.Flags))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	info := nmap.Detect(ctx, nmapPath)
	if info.OK {
		log.Printf("nmap: %s at %s", info.Version, info.Path)
	} else {
		log.Printf("nmap: NOT DETECTED (%s) — UI will show install instructions", info.Error)
	}

	priv := auth.Detect()
	log.Printf("privilege: %s (%s)", priv.Mode, priv.Reason)
	if privileged && !priv.AllowsRaw() {
		log.Printf("warning: --privileged was set but the process is not actually elevated; relaunch with `sudo n-mapped`")
	}

	hist, err := store.NewHistory(dataDir)
	if err != nil {
		return fmt.Errorf("opening history store: %w", err)
	}
	favs, err := store.NewFavorites(dataDir)
	if err != nil {
		return fmt.Errorf("opening favorites store: %w", err)
	}

	srv, err := server.New(server.Options{
		Bind:      bind,
		NmapPath:  nmapPath,
		Catalog:   cat,
		NmapInfo:  info,
		Privs:     priv,
		Runner:    nmap.NewRunner(),
		History:   hist,
		Favorites: favs,
	})
	if err != nil {
		return err
	}

	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()

	if !noBrowser {
		go openBrowser(srv.URL())
	}

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
		log.Println("shutting down...")
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutCtx)
	}
}
