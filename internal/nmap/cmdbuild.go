package nmap

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"sort"
	"strings"

	"github.com/nick-the-descended/n-mapped/internal/auth"
	"github.com/nick-the-descended/n-mapped/internal/catalog"
)

// Request is what the frontend POSTs to /api/scans. It contains only flag IDs
// (not raw argv) so user input can never inject arbitrary nmap arguments.
type Request struct {
	Targets    []string                  `json:"targets"`
	FlagIDs    []string                  `json:"flag_ids"`
	FlagValues map[string]string         `json:"flag_values,omitempty"` // flag_id -> stringified value for non-boolean flags
	ScriptIDs  []string                  `json:"script_ids,omitempty"`
	ScriptArgs map[string]map[string]any `json:"script_args,omitempty"`
}

// Built is the validated, materialized command ready to hand to os/exec.
type Built struct {
	Argv      []string `json:"argv"`        // includes the binary path at [0]
	Display   string   `json:"display"`     // shell-quoted form for the UI command preview
	NeedsRoot bool     `json:"needs_root"`
}

// Build validates a Request against the catalog and the current privilege
// State and produces the argv to exec. It refuses elevated-only flags when
// not elevated — the backend is authoritative; the UI can lie.
func Build(req Request, cat *catalog.Catalog, info Info, priv auth.State) (Built, error) {
	if len(req.Targets) == 0 {
		return Built{}, errors.New("at least one target is required")
	}
	for _, t := range req.Targets {
		if err := validateTarget(t); err != nil {
			return Built{}, fmt.Errorf("invalid target %q: %w", t, err)
		}
	}

	args := []string{}
	exclusiveSeen := map[string]string{} // exclusive-group -> first flag id seen
	needsRoot := false

	// Sort flag IDs so the output is deterministic across requests.
	flagIDs := append([]string(nil), req.FlagIDs...)
	sort.Strings(flagIDs)

	for _, id := range flagIDs {
		flag, ok := cat.Flag(id)
		if !ok {
			return Built{}, fmt.Errorf("unknown flag id: %s", id)
		}
		if flag.RequiresRoot {
			needsRoot = true
			if !priv.AllowsRaw() {
				return Built{}, fmt.Errorf("flag %s (%s) requires elevated privileges; relaunch with --privileged or `sudo n-mapped`", flag.ID, flag.Short)
			}
		}
		if flag.MinVersion != "" && info.OK {
			if !versionGTE(info, flag.MinVersion) {
				return Built{}, fmt.Errorf("flag %s requires nmap >= %s; you have %s", flag.ID, flag.MinVersion, info.Version)
			}
		}
		for _, group := range flag.MutuallyExclusiveWith {
			if other, seen := exclusiveSeen[group]; seen && other != flag.ID {
				return Built{}, fmt.Errorf("flag %s conflicts with %s", flag.ID, other)
			}
			exclusiveSeen[group] = flag.ID
		}

		token := flag.Short
		if token == "" {
			token = flag.Long
		}
		if token == "" {
			return Built{}, fmt.Errorf("catalog entry %s has no short or long form", flag.ID)
		}
		args = append(args, token)
		if flag.ValueType != "" && flag.ValueType != "boolean" {
			val, ok := req.FlagValues[flag.ID]
			if !ok || val == "" {
				return Built{}, fmt.Errorf("flag %s requires a value", flag.ID)
			}
			if err := validateValue(flag.ValueType, val); err != nil {
				return Built{}, fmt.Errorf("flag %s: %w", flag.ID, err)
			}
			args = append(args, val)
		}
	}

	// XML on stdout is mandatory — that's how the runner streams structured events.
	args = append(args, "-oX", "-")
	// Periodic progress lines so the UI can render an ETA.
	args = append(args, "--stats-every", "2s")
	// Targets last, after `--`, so a hostname starting with '-' can't be misread.
	args = append(args, "--")
	args = append(args, req.Targets...)

	bin := info.Path
	if bin == "" {
		bin = "nmap"
	}
	argv := append([]string{bin}, args...)
	return Built{Argv: argv, Display: shellQuote(argv), NeedsRoot: needsRoot}, nil
}

// validateTarget rejects shell metacharacters defensively. The argv path
// already prevents shell interpretation, but we keep targets to a tight
// allowlist of characters nmap itself accepts in target syntax.
func validateTarget(t string) error {
	if t == "" {
		return errors.New("empty")
	}
	if strings.ContainsAny(t, " \t\r\n;|&`$<>()'\"\\") {
		return errors.New("contains forbidden characters")
	}
	if strings.HasPrefix(t, "-") {
		return errors.New("must not start with '-'")
	}
	return nil
}

var (
	portListRE = regexp.MustCompile(`^[0-9TUSP,\-]+$`)
	intRE      = regexp.MustCompile(`^\d+$`)
)

func validateValue(kind, val string) error {
	switch kind {
	case "string":
		if strings.ContainsAny(val, "\n\r") {
			return errors.New("string value contains newline")
		}
		return nil
	case "int":
		if !intRE.MatchString(val) {
			return errors.New("expected integer")
		}
		return nil
	case "port-list":
		if !portListRE.MatchString(val) {
			return errors.New("expected port list (e.g. 22,80,443 or 1-1024 or T:80,U:53)")
		}
		return nil
	case "ip":
		if net.ParseIP(val) == nil {
			return errors.New("expected an IP address")
		}
		return nil
	default:
		return nil
	}
}

func versionGTE(info Info, want string) bool {
	parts := strings.SplitN(want, ".", 2)
	if len(parts) != 2 {
		return true
	}
	wm, wn := atoi(parts[0]), atoi(parts[1])
	return info.AtLeast(wm, wn)
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return n
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func shellQuote(argv []string) string {
	var b strings.Builder
	for i, a := range argv {
		if i > 0 {
			b.WriteByte(' ')
		}
		if needsQuote(a) {
			b.WriteByte('\'')
			b.WriteString(strings.ReplaceAll(a, "'", `'\''`))
			b.WriteByte('\'')
		} else {
			b.WriteString(a)
		}
	}
	return b.String()
}

func needsQuote(s string) bool {
	if s == "" {
		return true
	}
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			continue
		}
		switch r {
		case '-', '_', '.', '/', ':', ',', '=':
			continue
		}
		return true
	}
	return false
}
