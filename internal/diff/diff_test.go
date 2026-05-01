package diff

import "testing"

import "github.com/nick-the-descended/n-mapped/internal/nmap"

func host(addr string, status string, ports ...nmap.Port) nmap.Host {
	h := nmap.Host{
		Status:    nmap.Status{State: status},
		Addresses: []nmap.Address{{Addr: addr, AddrType: "ipv4"}},
	}
	if len(ports) > 0 {
		h.Ports = &nmap.Ports{Ports: ports}
	}
	return h
}

func tcpPort(id int, state, service, product, version string) nmap.Port {
	p := nmap.Port{Protocol: "tcp", PortID: id, State: nmap.PortState{State: state}}
	if service != "" || product != "" || version != "" {
		p.Service = &nmap.Service{Name: service, Product: product, Version: version}
	}
	return p
}

func TestCompare_Identical(t *testing.T) {
	a := &nmap.Run{Hosts: []nmap.Host{host("10.0.0.1", "up", tcpPort(22, "open", "ssh", "OpenSSH", "9.6"))}}
	b := &nmap.Run{Hosts: []nmap.Host{host("10.0.0.1", "up", tcpPort(22, "open", "ssh", "OpenSSH", "9.6"))}}
	r := Compare(a, b)
	if !r.Identical {
		t.Fatalf("expected identical, got: %+v", r)
	}
}

func TestCompare_HostAddedAndRemoved(t *testing.T) {
	a := &nmap.Run{Hosts: []nmap.Host{host("10.0.0.1", "up")}}
	b := &nmap.Run{Hosts: []nmap.Host{host("10.0.0.2", "up")}}
	r := Compare(a, b)
	if r.Identical {
		t.Fatal("expected differences")
	}
	if len(r.HostsRemoved) != 1 || r.HostsRemoved[0].Address != "10.0.0.1" {
		t.Errorf("removed: %+v", r.HostsRemoved)
	}
	if len(r.HostsAdded) != 1 || r.HostsAdded[0].Address != "10.0.0.2" {
		t.Errorf("added: %+v", r.HostsAdded)
	}
}

func TestCompare_PortChanges(t *testing.T) {
	a := &nmap.Run{Hosts: []nmap.Host{host("10.0.0.1", "up",
		tcpPort(22, "open", "ssh", "OpenSSH", "9.5"),
		tcpPort(80, "open", "http", "nginx", "1.26"),
	)}}
	b := &nmap.Run{Hosts: []nmap.Host{host("10.0.0.1", "up",
		tcpPort(22, "open", "ssh", "OpenSSH", "9.6"), // version changed
		tcpPort(443, "open", "https", "nginx", "1.27"), // newly open
		// 80 removed
	)}}
	r := Compare(a, b)
	if r.Identical {
		t.Fatal("expected differences")
	}
	if len(r.HostsChanged) != 1 {
		t.Fatalf("expected 1 changed host, got %d", len(r.HostsChanged))
	}
	hc := r.HostsChanged[0]
	if len(hc.PortsAdded) != 1 || hc.PortsAdded[0].PortID != 443 {
		t.Errorf("ports_added: %+v", hc.PortsAdded)
	}
	if len(hc.PortsRemoved) != 1 || hc.PortsRemoved[0].PortID != 80 {
		t.Errorf("ports_removed: %+v", hc.PortsRemoved)
	}
	if len(hc.PortsChanged) != 1 || hc.PortsChanged[0].PortID != 22 {
		t.Fatalf("ports_changed: %+v", hc.PortsChanged)
	}
	pc := hc.PortsChanged[0]
	if pc.VersionBefore != "OpenSSH 9.5" || pc.VersionAfter != "OpenSSH 9.6" {
		t.Errorf("version diff: %q -> %q", pc.VersionBefore, pc.VersionAfter)
	}
}

func TestCompare_StatusChange(t *testing.T) {
	a := &nmap.Run{Hosts: []nmap.Host{host("10.0.0.1", "up")}}
	b := &nmap.Run{Hosts: []nmap.Host{host("10.0.0.1", "down")}}
	r := Compare(a, b)
	if len(r.HostsChanged) != 1 {
		t.Fatalf("expected 1 changed host: %+v", r)
	}
	if r.HostsChanged[0].StatusBefore != "up" || r.HostsChanged[0].StatusAfter != "down" {
		t.Errorf("status diff: %+v", r.HostsChanged[0])
	}
}
