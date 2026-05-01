// Package store persists scan history. Phase 2 uses JSON files in the data
// directory (one per scan); Phase 7 will migrate to SQLite when we need
// asset tracking and time-series queries. The on-disk shape is the public
// contract — keep it backward-compatible across migrations.
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/nick-the-descended/n-mapped/internal/nmap"
)

// Record is the persisted history entry for a single scan.
type Record struct {
	ID         string            `json:"id"`
	Display    string            `json:"display"`
	Argv       []string          `json:"argv"`
	Targets    []string          `json:"targets"`
	FlagIDs    []string          `json:"flag_ids,omitempty"`
	ScriptIDs  []string          `json:"script_ids,omitempty"`
	ScriptArgs map[string]string `json:"script_args,omitempty"`
	Started    time.Time         `json:"started"`
	Ended      time.Time         `json:"ended"`
	ExitCode   int               `json:"exit_code"`
	Result     *nmap.Result      `json:"result,omitempty"`
	Error      string            `json:"error,omitempty"`
	Tags       []string          `json:"tags,omitempty"`
	Notes      string            `json:"notes,omitempty"`
}

// Summary is a slimmed Record used for the history list (no embedded
// XML / parsed run, so listing 1000s of scans stays fast).
type Summary struct {
	ID         string    `json:"id"`
	Display    string    `json:"display"`
	Targets    []string  `json:"targets"`
	Started    time.Time `json:"started"`
	Ended      time.Time `json:"ended"`
	ExitCode   int       `json:"exit_code"`
	HostsUp    int       `json:"hosts_up"`
	HostsTotal int       `json:"hosts_total"`
	OpenPorts  int       `json:"open_ports"`
	Tags       []string  `json:"tags,omitempty"`
	Notes      string    `json:"notes,omitempty"`
	Error      string    `json:"error,omitempty"`
}

// History is a JSON-file-backed history store. Safe for concurrent use.
type History struct {
	dir string
	mu  sync.Mutex
}

// NewHistory creates the history directory under dataDir/history and returns
// a History rooted there.
func NewHistory(dataDir string) (*History, error) {
	dir := filepath.Join(dataDir, "history")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("creating history dir: %w", err)
	}
	return &History{dir: dir}, nil
}

// idPattern restricts IDs to the hex form produced by runner.newScanID(),
// so the path-join of dir + id can never escape dir.
var idPattern = regexp.MustCompile(`^[a-f0-9]{8,64}$`)

// Save writes a record atomically (tmp + rename). Existing entries are
// overwritten — last-write-wins.
func (h *History) Save(rec Record) error {
	if !idPattern.MatchString(rec.ID) {
		return fmt.Errorf("invalid scan id: %q", rec.ID)
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	path := filepath.Join(h.dir, rec.ID+".json")
	tmp := path + ".tmp"
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Get loads a record by ID.
func (h *History) Get(id string) (*Record, error) {
	if !idPattern.MatchString(id) {
		return nil, errors.New("invalid scan id")
	}
	data, err := os.ReadFile(filepath.Join(h.dir, id+".json"))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var rec Record
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

// Delete removes a record by ID. Returns nil if it didn't exist.
func (h *History) Delete(id string) error {
	if !idPattern.MatchString(id) {
		return errors.New("invalid scan id")
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	err := os.Remove(filepath.Join(h.dir, id+".json"))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

// List returns summaries newest-first. Best-effort: bad files are skipped.
func (h *History) List(limit int) ([]Summary, error) {
	entries, err := os.ReadDir(h.dir)
	if err != nil {
		return nil, err
	}
	out := make([]Summary, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		id := e.Name()[:len(e.Name())-len(".json")]
		rec, err := h.Get(id)
		if err != nil || rec == nil {
			continue
		}
		out = append(out, summarize(rec))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Started.After(out[j].Started) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// UpdateMeta replaces the tags and notes on an existing record. Returns an
// error if the record doesn't exist.
func (h *History) UpdateMeta(id string, tags []string, notes string) error {
	if !idPattern.MatchString(id) {
		return errors.New("invalid scan id")
	}
	rec, err := h.Get(id)
	if err != nil {
		return err
	}
	if rec == nil {
		return errors.New("not found")
	}
	rec.Tags = tags
	rec.Notes = notes
	return h.Save(*rec)
}

// RecordFromResult builds a Record from a runner.Result. The runner already
// carries the source request metadata (targets / flag_ids / script_ids).
func RecordFromResult(res nmap.Result) Record {
	return Record{
		ID:         res.ID,
		Display:    res.Display,
		Argv:       res.Argv,
		Targets:    res.Targets,
		FlagIDs:    res.FlagIDs,
		ScriptIDs:  res.ScriptIDs,
		ScriptArgs: res.ScriptArgs,
		Started:    res.Started,
		Ended:      res.Ended,
		ExitCode:   res.ExitCode,
		Result:     &res,
		Error:      res.Error,
	}
}

func summarize(r *Record) Summary {
	s := Summary{
		ID:       r.ID,
		Display:  r.Display,
		Targets:  r.Targets,
		Started:  r.Started,
		Ended:    r.Ended,
		ExitCode: r.ExitCode,
		Tags:     r.Tags,
		Notes:    r.Notes,
		Error:    r.Error,
	}
	if r.Result != nil && r.Result.Run != nil {
		run := r.Result.Run
		if run.RunStats != nil {
			s.HostsUp = run.RunStats.Hosts.Up
			s.HostsTotal = run.RunStats.Hosts.Total
		}
		for _, host := range run.Hosts {
			if host.Ports == nil {
				continue
			}
			for _, p := range host.Ports.Ports {
				if p.State.State == "open" {
					s.OpenPorts++
				}
			}
		}
	}
	return s
}
