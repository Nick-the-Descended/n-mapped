package nmap

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Info describes the locally installed nmap binary.
type Info struct {
	Path    string `json:"path"`
	Version string `json:"version"`        // e.g. "7.94"
	Major   int    `json:"major"`
	Minor   int    `json:"minor"`
	Raw     string `json:"raw"`            // first line of `nmap --version`
	OK      bool   `json:"ok"`
	Error   string `json:"error,omitempty"`
}

var versionRE = regexp.MustCompile(`Nmap version (\d+)\.(\d+)`)

// Detect runs `<path> --version` and parses the output. Returns an Info with
// OK=false (and Error populated) rather than failing the caller — the UI can
// then show a friendly "install nmap" page.
func Detect(ctx context.Context, path string) Info {
	if path == "" {
		path = "nmap"
	}
	resolved, err := exec.LookPath(path)
	if err != nil {
		return Info{Path: path, Error: fmt.Sprintf("nmap not found in PATH: %v", err)}
	}
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(cctx, resolved, "--version").Output()
	if err != nil {
		return Info{Path: resolved, Error: fmt.Sprintf("running nmap --version: %v", err)}
	}
	first := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]
	info := Info{Path: resolved, Raw: first}
	m := versionRE.FindStringSubmatch(first)
	if m == nil {
		info.Error = "could not parse version from: " + first
		return info
	}
	info.Major, _ = strconv.Atoi(m[1])
	info.Minor, _ = strconv.Atoi(m[2])
	info.Version = fmt.Sprintf("%d.%d", info.Major, info.Minor)
	info.OK = true
	return info
}

// AtLeast reports whether the detected version is >= the given major.minor.
// Returns false if Detect failed.
func (i Info) AtLeast(major, minor int) bool {
	if !i.OK {
		return false
	}
	if i.Major != major {
		return i.Major > major
	}
	return i.Minor >= minor
}
