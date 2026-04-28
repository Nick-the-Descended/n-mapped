package store

import (
	"crypto/rand"
	"encoding/hex"
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
)

// Favorite is a saved builder template the user can re-apply with one click.
type Favorite struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Targets     string            `json:"targets,omitempty"`
	FlagIDs     []string          `json:"flag_ids,omitempty"`
	FlagValues  map[string]string `json:"flag_values,omitempty"`
	ScriptIDs   []string          `json:"script_ids,omitempty"`
	ScriptArgs  map[string]string `json:"script_args,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// Favorites is a JSON-file-backed favorites store.
type Favorites struct {
	dir string
	mu  sync.Mutex
}

// NewFavorites creates the favorites directory under dataDir/favorites.
func NewFavorites(dataDir string) (*Favorites, error) {
	dir := filepath.Join(dataDir, "favorites")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("creating favorites dir: %w", err)
	}
	return &Favorites{dir: dir}, nil
}

var favIDPattern = regexp.MustCompile(`^[a-f0-9]{8,64}$`)

// NewID generates a fresh favorite id.
func newFavoriteID() string {
	var b [10]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// Save creates or replaces a favorite. Generates an ID and timestamps if absent.
func (f *Favorites) Save(fav Favorite) (Favorite, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if fav.Name == "" {
		return Favorite{}, errors.New("name is required")
	}
	if fav.ID == "" {
		fav.ID = newFavoriteID()
		fav.CreatedAt = time.Now()
	} else if !favIDPattern.MatchString(fav.ID) {
		return Favorite{}, fmt.Errorf("invalid favorite id: %q", fav.ID)
	}
	fav.UpdatedAt = time.Now()
	if fav.CreatedAt.IsZero() {
		fav.CreatedAt = fav.UpdatedAt
	}

	data, err := json.MarshalIndent(fav, "", "  ")
	if err != nil {
		return Favorite{}, err
	}
	path := filepath.Join(f.dir, fav.ID+".json")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return Favorite{}, err
	}
	if err := os.Rename(tmp, path); err != nil {
		return Favorite{}, err
	}
	return fav, nil
}

// Get loads a favorite by ID.
func (f *Favorites) Get(id string) (*Favorite, error) {
	if !favIDPattern.MatchString(id) {
		return nil, errors.New("invalid favorite id")
	}
	data, err := os.ReadFile(filepath.Join(f.dir, id+".json"))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var fav Favorite
	if err := json.Unmarshal(data, &fav); err != nil {
		return nil, err
	}
	return &fav, nil
}

// Delete removes a favorite. Returns nil if it didn't exist.
func (f *Favorites) Delete(id string) error {
	if !favIDPattern.MatchString(id) {
		return errors.New("invalid favorite id")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	err := os.Remove(filepath.Join(f.dir, id+".json"))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

// List returns all favorites newest-first.
func (f *Favorites) List() ([]Favorite, error) {
	entries, err := os.ReadDir(f.dir)
	if err != nil {
		return nil, err
	}
	out := make([]Favorite, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		id := e.Name()[:len(e.Name())-len(".json")]
		fav, err := f.Get(id)
		if err != nil || fav == nil {
			continue
		}
		out = append(out, *fav)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out, nil
}
