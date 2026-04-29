// Package update polls the GitHub Releases API once an hour for newer
// versions of n-mapped and exposes the result via Status(). The check is
// disabled when the binary is built without a release version (i.e. dev
// builds reporting "dev"), and can be turned off by the user with the
// --no-update-check launch flag.
package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

const releasesURL = "https://api.github.com/repos/nick-the-descended/n-mapped/releases/latest"

// Status is the result surfaced to the UI.
type Status struct {
	CurrentVersion string    `json:"current_version"`
	LatestVersion  string    `json:"latest_version,omitempty"`
	HasUpdate      bool      `json:"has_update"`
	URL            string    `json:"url,omitempty"`
	CheckedAt      time.Time `json:"checked_at,omitempty"`
	Enabled        bool      `json:"enabled"`
	Error          string    `json:"error,omitempty"`
}

// Checker holds the latest cached check result.
type Checker struct {
	current string
	enabled bool
	mu      sync.RWMutex
	status  Status
}

// New returns a Checker. If currentVersion == "dev" or "" the checker stays
// disabled, since dev builds shouldn't nag.
func New(currentVersion string, enabled bool) *Checker {
	if currentVersion == "" || currentVersion == "dev" {
		enabled = false
	}
	c := &Checker{current: currentVersion, enabled: enabled}
	c.status = Status{CurrentVersion: currentVersion, Enabled: enabled}
	return c
}

// Status returns a copy of the current cached status.
func (c *Checker) Status() Status {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status
}

// Run starts the background poller. Returns immediately; cancel ctx to stop.
func (c *Checker) Run(ctx context.Context) {
	if !c.enabled {
		return
	}
	go func() {
		// Stagger the first check by 5 seconds so it never blocks startup.
		t := time.NewTimer(5 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				c.checkOnce(ctx)
				t.Reset(time.Hour)
			}
		}
	}()
}

func (c *Checker) checkOnce(ctx context.Context) {
	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(cctx, http.MethodGet, releasesURL, nil)
	if err != nil {
		c.recordError(err)
		return
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "n-mapped/"+c.current)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.recordError(err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		c.recordError(fmt.Errorf("github api: %s", resp.Status))
		return
	}
	var body struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		c.recordError(err)
		return
	}
	latest := strings.TrimPrefix(body.TagName, "v")
	current := strings.TrimPrefix(c.current, "v")
	c.mu.Lock()
	c.status = Status{
		CurrentVersion: c.current,
		LatestVersion:  body.TagName,
		HasUpdate:      latest != "" && current != "" && newer(latest, current),
		URL:            body.HTMLURL,
		CheckedAt:      time.Now(),
		Enabled:        true,
	}
	c.mu.Unlock()
}

func (c *Checker) recordError(err error) {
	c.mu.Lock()
	c.status = Status{
		CurrentVersion: c.current,
		CheckedAt:      time.Now(),
		Enabled:        true,
		Error:          err.Error(),
	}
	c.mu.Unlock()
}

// newer reports whether a > b, treating both as dotted version strings.
// Trailing pre-release tags ("-rc1") are ignored. Returns false on parse fail.
func newer(a, b string) bool {
	pa := splitVersion(a)
	pb := splitVersion(b)
	for i := 0; i < len(pa) || i < len(pb); i++ {
		var ai, bi int
		if i < len(pa) {
			ai = pa[i]
		}
		if i < len(pb) {
			bi = pb[i]
		}
		if ai != bi {
			return ai > bi
		}
	}
	return false
}

func splitVersion(s string) []int {
	if i := strings.IndexAny(s, "-+"); i >= 0 {
		s = s[:i]
	}
	parts := strings.Split(s, ".")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		n := 0
		for _, c := range p {
			if c < '0' || c > '9' {
				return out
			}
			n = n*10 + int(c-'0')
		}
		out = append(out, n)
	}
	return out
}
