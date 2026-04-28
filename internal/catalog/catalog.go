// Package catalog loads the curated catalog of nmap flags (and, in later
// phases, NSE scripts) that drives the UI. The catalog is embedded into the
// binary as JSON; entries are sourced from human-curated YAML in catalog-data/
// and built by tools/build-catalog.
package catalog

import (
	"embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

//go:embed data/flags.json data/scripts.json
var embedded embed.FS

// Flag describes one nmap flag the UI exposes.
type Flag struct {
	ID                    string     `json:"id"`
	Short                 string     `json:"short,omitempty"` // -sS
	Long                  string     `json:"long,omitempty"`  // --open
	Category              string     `json:"category"`
	ValueType             string     `json:"value_type,omitempty"` // boolean | string | int | port-list | ip
	SkillLevel            string     `json:"skill_level"`          // beginner | intermediate | advanced
	RequiresRoot          bool       `json:"requires_root,omitempty"`
	MinVersion            string     `json:"min_version,omitempty"`
	MaxVersion            string     `json:"max_version,omitempty"`
	MutuallyExclusiveWith []string   `json:"mutually_exclusive_with,omitempty"`
	Implies               []string   `json:"implies,omitempty"`
	ShortDescription      string     `json:"short_description"`
	LongDescription       string     `json:"long_description,omitempty"`
	Examples              []Example  `json:"examples,omitempty"`
	Warnings              []Warning  `json:"warnings,omitempty"`
	References            []Refer    `json:"references,omitempty"`
	Tags                  []string   `json:"tags,omitempty"`
}

type Example struct {
	Command     string `json:"command"`
	Explanation string `json:"explanation,omitempty"`
}

type Warning struct {
	Level string `json:"level"` // info | warn | danger
	Text  string `json:"text"`
}

type Refer struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

// Category is a top-level grouping in the UI sidebar.
type Category struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Order       int     `json:"order"`
}

type fileFormat struct {
	SchemaVersion string     `json:"schema_version"`
	Categories    []Category `json:"categories"`
	Flags         []Flag     `json:"flags"`
}

type scriptsFile struct {
	SchemaVersion string   `json:"schema_version"`
	Scripts       []Script `json:"scripts"`
}

// Script describes one NSE script the UI exposes. Mirrors the flag schema
// where it makes sense, but adds NSE-specific fields (categories, args).
type Script struct {
	ID               string      `json:"id"`
	Categories       []string    `json:"categories"`
	SkillLevel       string      `json:"skill_level"`
	ShortDescription string      `json:"short_description"`
	LongDescription  string      `json:"long_description,omitempty"`
	Args             []ScriptArg `json:"args,omitempty"`
	Examples         []Example   `json:"examples,omitempty"`
	Warnings         []Warning   `json:"warnings,omitempty"`
	References       []Refer     `json:"references,omitempty"`
	Tags             []string    `json:"tags,omitempty"`
}

// ScriptArg describes a single --script-args parameter.
type ScriptArg struct {
	Name        string `json:"name"`        // e.g. "dns-brute.threads"
	Type        string `json:"type"`        // string | int | boolean
	Default     string `json:"default,omitempty"`
	Description string `json:"description,omitempty"`
}

// Catalog is the in-memory catalog used by the rest of the app.
type Catalog struct {
	SchemaVersion string
	Categories    []Category
	Flags         []Flag
	Scripts       []Script
	byID          map[string]*Flag
	byScriptID    map[string]*Script
}

// Load reads the embedded catalog and returns a queryable Catalog.
func Load() (*Catalog, error) {
	data, err := embedded.ReadFile("data/flags.json")
	if err != nil {
		return nil, fmt.Errorf("reading embedded flags catalog: %w", err)
	}
	var f fileFormat
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parsing flags catalog: %w", err)
	}

	scriptsData, err := embedded.ReadFile("data/scripts.json")
	if err != nil {
		return nil, fmt.Errorf("reading embedded scripts catalog: %w", err)
	}
	var sf scriptsFile
	if err := json.Unmarshal(scriptsData, &sf); err != nil {
		return nil, fmt.Errorf("parsing scripts catalog: %w", err)
	}

	c := &Catalog{
		SchemaVersion: f.SchemaVersion,
		Categories:    f.Categories,
		Flags:         f.Flags,
		Scripts:       sf.Scripts,
		byID:          make(map[string]*Flag, len(f.Flags)),
		byScriptID:    make(map[string]*Script, len(sf.Scripts)),
	}
	sort.SliceStable(c.Categories, func(i, j int) bool { return c.Categories[i].Order < c.Categories[j].Order })
	for i := range c.Flags {
		c.byID[c.Flags[i].ID] = &c.Flags[i]
	}
	for i := range c.Scripts {
		c.byScriptID[c.Scripts[i].ID] = &c.Scripts[i]
	}
	return c, nil
}

// Script looks up an NSE script by ID.
func (c *Catalog) Script(id string) (*Script, bool) {
	s, ok := c.byScriptID[id]
	return s, ok
}

// Flag looks up a flag by ID.
func (c *Catalog) Flag(id string) (*Flag, bool) {
	f, ok := c.byID[id]
	return f, ok
}

// Search returns flags whose ID, short/long form, description, or tags match
// the query (case-insensitive substring). Phase 3 will replace this with a
// proper fuzzy matcher.
func (c *Catalog) Search(query string) []Flag {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return c.Flags
	}
	var out []Flag
	for _, f := range c.Flags {
		if matches(f, q) {
			out = append(out, f)
		}
	}
	return out
}

func matches(f Flag, q string) bool {
	for _, hay := range []string{f.ID, f.Short, f.Long, f.ShortDescription, f.LongDescription, f.Category} {
		if strings.Contains(strings.ToLower(hay), q) {
			return true
		}
	}
	for _, t := range f.Tags {
		if strings.Contains(strings.ToLower(t), q) {
			return true
		}
	}
	return false
}
