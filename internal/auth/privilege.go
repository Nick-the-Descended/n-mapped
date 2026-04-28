// Package auth detects whether n-mapped has the privileges nmap needs for
// raw-socket scans (-sS, -sU, -O, etc.).
//
// Privilege is decided once at app launch — by EUID, by a Linux capability set
// on the nmap binary, or by the user passing --privileged (which can re-exec
// via pkexec on Linux). Per-scan password prompts are never used.
package auth

import (
	"os"
	"runtime"
	"strconv"
	"strings"
)

// Mode describes how the running process can perform raw-socket scans.
type Mode string

const (
	ModeUser       Mode = "user"       // unprivileged, raw-socket scans disabled
	ModeCapability Mode = "capability" // Linux file caps grant cap_net_raw/admin
	ModeRoot       Mode = "root"       // EUID == 0
)

// State summarizes the launch-time privilege context. Surfaced at /api/privilege.
type State struct {
	Mode       Mode   `json:"mode"`
	EUID       int    `json:"euid"`
	OS         string `json:"os"`
	Capability string `json:"capability,omitempty"` // raw CapEff hex string when read
	Reason     string `json:"reason,omitempty"`     // human-readable context
}

// Detect inspects the current process and returns the privilege State.
//
// Note: capability detection here only describes the *running* process. nmap
// itself may carry file-caps independent of this; that is reported separately
// by internal/nmap.Version().
func Detect() State {
	euid := os.Geteuid()
	st := State{EUID: euid, OS: runtime.GOOS}
	if euid == 0 {
		st.Mode = ModeRoot
		st.Reason = "running as root"
		return st
	}
	if runtime.GOOS == "linux" {
		if hasNetRaw, capHex := readEffectiveCaps(); hasNetRaw {
			st.Mode = ModeCapability
			st.Capability = capHex
			st.Reason = "process has cap_net_raw"
			return st
		}
	}
	st.Mode = ModeUser
	st.Reason = "unprivileged user; raw-socket scans disabled"
	return st
}

// AllowsRaw reports whether the current State permits raw-socket nmap scans.
func (s State) AllowsRaw() bool {
	return s.Mode == ModeRoot || s.Mode == ModeCapability
}

// readEffectiveCaps parses /proc/self/status on Linux and returns whether
// cap_net_raw (bit 13) is in CapEff, plus the raw hex string for display.
func readEffectiveCaps() (bool, string) {
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return false, ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "CapEff:") {
			continue
		}
		hex := strings.TrimSpace(strings.TrimPrefix(line, "CapEff:"))
		caps, err := strconv.ParseUint(hex, 16, 64)
		if err != nil {
			return false, hex
		}
		const capNetRaw = 13
		return caps&(1<<capNetRaw) != 0, hex
	}
	return false, ""
}
