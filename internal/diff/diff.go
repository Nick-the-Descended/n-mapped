// Package diff computes the structured difference between two nmap scan
// results. Used by the Diff tab in the UI and by the (later) recurring
// scheduler's "alert on change" feature.
package diff

import (
	"sort"

	"github.com/nick-the-descended/n-mapped/internal/nmap"
)

// PortKey identifies a port within a host across runs.
type PortKey struct {
	Protocol string
	PortID   int
}

// PortInfo is the slim view of a port used in diffs.
type PortInfo struct {
	Protocol string `json:"protocol"`
	PortID   int    `json:"portid"`
	State    string `json:"state"`
	Service  string `json:"service,omitempty"`
	Product  string `json:"product,omitempty"`
	Version  string `json:"version,omitempty"`
}

// PortChange describes how one port differs between two scans.
type PortChange struct {
	Protocol      string `json:"protocol"`
	PortID        int    `json:"portid"`
	StateBefore   string `json:"state_before,omitempty"`
	StateAfter    string `json:"state_after,omitempty"`
	ServiceBefore string `json:"service_before,omitempty"`
	ServiceAfter  string `json:"service_after,omitempty"`
	VersionBefore string `json:"version_before,omitempty"`
	VersionAfter  string `json:"version_after,omitempty"`
}

// HostSummary identifies a host briefly.
type HostSummary struct {
	Address  string `json:"address"`
	Hostname string `json:"hostname,omitempty"`
}

// HostChange describes how a host common to both scans differs.
type HostChange struct {
	Address      string       `json:"address"`
	Hostname     string       `json:"hostname,omitempty"`
	StatusBefore string       `json:"status_before,omitempty"`
	StatusAfter  string       `json:"status_after,omitempty"`
	PortsAdded   []PortInfo   `json:"ports_added,omitempty"`
	PortsRemoved []PortInfo   `json:"ports_removed,omitempty"`
	PortsChanged []PortChange `json:"ports_changed,omitempty"`
	OSBefore     string       `json:"os_before,omitempty"`
	OSAfter      string       `json:"os_after,omitempty"`
}

// Result is the diff between two nmap.Run snapshots.
type Result struct {
	HostsAdded   []HostSummary `json:"hosts_added,omitempty"`
	HostsRemoved []HostSummary `json:"hosts_removed,omitempty"`
	HostsChanged []HostChange  `json:"hosts_changed,omitempty"`
	Identical    bool          `json:"identical"`
}

// Compare returns the diff a → b. nil inputs are treated as empty runs.
func Compare(a, b *nmap.Run) Result {
	indexA := indexHosts(a)
	indexB := indexHosts(b)

	var out Result

	for addr, ha := range indexA {
		hb, ok := indexB[addr]
		if !ok {
			out.HostsRemoved = append(out.HostsRemoved, summarize(ha))
			continue
		}
		change, changed := compareHost(ha, hb)
		if changed {
			out.HostsChanged = append(out.HostsChanged, change)
		}
	}
	for addr, hb := range indexB {
		if _, ok := indexA[addr]; !ok {
			out.HostsAdded = append(out.HostsAdded, summarize(hb))
		}
	}

	sort.Slice(out.HostsAdded,   func(i, j int) bool { return out.HostsAdded[i].Address < out.HostsAdded[j].Address })
	sort.Slice(out.HostsRemoved, func(i, j int) bool { return out.HostsRemoved[i].Address < out.HostsRemoved[j].Address })
	sort.Slice(out.HostsChanged, func(i, j int) bool { return out.HostsChanged[i].Address < out.HostsChanged[j].Address })

	out.Identical = len(out.HostsAdded) == 0 && len(out.HostsRemoved) == 0 && len(out.HostsChanged) == 0
	return out
}

func indexHosts(r *nmap.Run) map[string]nmap.Host {
	out := map[string]nmap.Host{}
	if r == nil {
		return out
	}
	for _, h := range r.Hosts {
		addr := h.PrimaryAddress()
		if addr == "" {
			continue
		}
		out[addr] = h
	}
	return out
}

func summarize(h nmap.Host) HostSummary {
	s := HostSummary{Address: h.PrimaryAddress()}
	if len(h.Hostnames.Hostnames) > 0 {
		s.Hostname = h.Hostnames.Hostnames[0].Name
	}
	return s
}

func compareHost(a, b nmap.Host) (HostChange, bool) {
	change := HostChange{Address: a.PrimaryAddress()}
	if len(a.Hostnames.Hostnames) > 0 {
		change.Hostname = a.Hostnames.Hostnames[0].Name
	} else if len(b.Hostnames.Hostnames) > 0 {
		change.Hostname = b.Hostnames.Hostnames[0].Name
	}
	changed := false

	if a.Status.State != b.Status.State {
		change.StatusBefore = a.Status.State
		change.StatusAfter = b.Status.State
		changed = true
	}

	pa := indexPorts(a)
	pb := indexPorts(b)

	for k, port := range pa {
		other, ok := pb[k]
		if !ok {
			change.PortsRemoved = append(change.PortsRemoved, portInfo(port))
			changed = true
			continue
		}
		pc, diff := comparePort(port, other)
		if diff {
			change.PortsChanged = append(change.PortsChanged, pc)
			changed = true
		}
	}
	for k, port := range pb {
		if _, ok := pa[k]; !ok {
			change.PortsAdded = append(change.PortsAdded, portInfo(port))
			changed = true
		}
	}

	osA := osName(a)
	osB := osName(b)
	if osA != osB {
		change.OSBefore = osA
		change.OSAfter = osB
		changed = true
	}

	sort.Slice(change.PortsAdded,   func(i, j int) bool { return portKeyLess(change.PortsAdded[i], change.PortsAdded[j]) })
	sort.Slice(change.PortsRemoved, func(i, j int) bool { return portKeyLess(change.PortsRemoved[i], change.PortsRemoved[j]) })
	sort.Slice(change.PortsChanged, func(i, j int) bool {
		if change.PortsChanged[i].Protocol != change.PortsChanged[j].Protocol {
			return change.PortsChanged[i].Protocol < change.PortsChanged[j].Protocol
		}
		return change.PortsChanged[i].PortID < change.PortsChanged[j].PortID
	})
	return change, changed
}

func indexPorts(h nmap.Host) map[PortKey]nmap.Port {
	out := map[PortKey]nmap.Port{}
	if h.Ports == nil {
		return out
	}
	for _, p := range h.Ports.Ports {
		out[PortKey{Protocol: p.Protocol, PortID: p.PortID}] = p
	}
	return out
}

func portInfo(p nmap.Port) PortInfo {
	pi := PortInfo{Protocol: p.Protocol, PortID: p.PortID, State: p.State.State}
	if p.Service != nil {
		pi.Service = p.Service.Name
		pi.Product = p.Service.Product
		pi.Version = p.Service.Version
	}
	return pi
}

func comparePort(a, b nmap.Port) (PortChange, bool) {
	pc := PortChange{Protocol: a.Protocol, PortID: a.PortID}
	diff := false
	if a.State.State != b.State.State {
		pc.StateBefore = a.State.State
		pc.StateAfter = b.State.State
		diff = true
	}
	sa, sb := serviceName(a), serviceName(b)
	if sa != sb {
		pc.ServiceBefore = sa
		pc.ServiceAfter = sb
		diff = true
	}
	va, vb := serviceVersion(a), serviceVersion(b)
	if va != vb {
		pc.VersionBefore = va
		pc.VersionAfter = vb
		diff = true
	}
	return pc, diff
}

func serviceName(p nmap.Port) string {
	if p.Service == nil {
		return ""
	}
	return p.Service.Name
}

func serviceVersion(p nmap.Port) string {
	if p.Service == nil {
		return ""
	}
	if p.Service.Product == "" && p.Service.Version == "" {
		return ""
	}
	if p.Service.Version == "" {
		return p.Service.Product
	}
	return p.Service.Product + " " + p.Service.Version
}

func osName(h nmap.Host) string {
	if h.OS == nil || len(h.OS.Matches) == 0 {
		return ""
	}
	return h.OS.Matches[0].Name
}

func portKeyLess(a, b PortInfo) bool {
	if a.Protocol != b.Protocol {
		return a.Protocol < b.Protocol
	}
	return a.PortID < b.PortID
}
