package nmap

import (
	"strings"
	"testing"
)

// Minimal hand-crafted nmap XML sample covering the elements we care about.
// Modeled after a real `nmap -sV -O -oX -` run.
const sampleXML = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE nmaprun>
<nmaprun scanner="nmap" args="nmap -sV -O -oX - 192.0.2.5" start="1700000000" startstr="Tue Nov 14 12:33:20 2023" version="7.94" xmloutputversion="1.05">
<scaninfo type="syn" protocol="tcp" numservices="1000" services="1-1000"/>
<verbose level="0"/>
<host starttime="1700000001" endtime="1700000003">
  <status state="up" reason="echo-reply" reason_ttl="63"/>
  <address addr="192.0.2.5" addrtype="ipv4"/>
  <hostnames>
    <hostname name="example.test" type="user"/>
  </hostnames>
  <ports>
    <extraports state="closed" count="998"/>
    <port protocol="tcp" portid="22">
      <state state="open" reason="syn-ack" reason_ttl="63"/>
      <service name="ssh" product="OpenSSH" version="9.6p1" method="probed" conf="10"/>
    </port>
    <port protocol="tcp" portid="80">
      <state state="open" reason="syn-ack" reason_ttl="63"/>
      <service name="http" product="nginx" version="1.27.0"/>
      <script id="http-title" output="Welcome"/>
    </port>
  </ports>
  <os>
    <osmatch name="Linux 5.0 - 6.0" accuracy="98">
      <osclass type="general purpose" vendor="Linux" osfamily="Linux" osgen="5.X" accuracy="98"/>
    </osmatch>
  </os>
  <times srtt="40123" rttvar="2000" to="100000"/>
</host>
<taskprogress task="Service scan" time="1700000002" percent="42.5" remaining="3" etc="1700000005"/>
<runstats>
  <finished time="1700000003" timestr="Tue Nov 14 12:33:23 2023" elapsed="3.10" summary="Nmap done" exit="success"/>
  <hosts up="1" down="0" total="1"/>
</runstats>
</nmaprun>
`

func TestParseStream(t *testing.T) {
	var (
		gotRunStart bool
		gotScanInfo *ScanInfo
		hosts       []Host
		progress    *TaskProgress
		gotStats    *RunStats
	)
	cb := Callbacks{
		OnRunStart: func(scanner, args, version string, start int64) {
			gotRunStart = true
			if scanner != "nmap" || version != "7.94" || start != 1700000000 {
				t.Errorf("RunStart unexpected: scanner=%q version=%q start=%d", scanner, version, start)
			}
		},
		OnScanInfo:     func(s ScanInfo) { gotScanInfo = &s },
		OnHost:         func(h Host) { hosts = append(hosts, h) },
		OnTaskProgress: func(p TaskProgress) { progress = &p },
		OnRunStats:     func(rs RunStats) { gotStats = &rs },
	}

	raw, err := ParseStream(strings.NewReader(sampleXML), cb)
	if err != nil {
		t.Fatalf("ParseStream returned error: %v", err)
	}
	if len(raw) == 0 {
		t.Fatal("raw bytes were not captured")
	}
	if !gotRunStart {
		t.Error("OnRunStart was not invoked")
	}
	if gotScanInfo == nil || gotScanInfo.Type != "syn" {
		t.Errorf("scaninfo not captured: %+v", gotScanInfo)
	}
	if len(hosts) != 1 {
		t.Fatalf("expected 1 host, got %d", len(hosts))
	}
	h := hosts[0]
	if h.Status.State != "up" {
		t.Errorf("host status: %q", h.Status.State)
	}
	if h.PrimaryAddress() != "192.0.2.5" {
		t.Errorf("primary address: %q", h.PrimaryAddress())
	}
	if h.Ports == nil || len(h.Ports.Ports) != 2 {
		t.Fatalf("expected 2 ports, got %#v", h.Ports)
	}
	p22 := h.Ports.Ports[0]
	if p22.PortID != 22 || p22.State.State != "open" || p22.Service == nil || p22.Service.Name != "ssh" {
		t.Errorf("unexpected port 22: %+v / %+v", p22, p22.Service)
	}
	p80 := h.Ports.Ports[1]
	if p80.Service == nil || p80.Service.Product != "nginx" {
		t.Errorf("port 80 service: %+v", p80.Service)
	}
	if len(p80.Scripts) != 1 || p80.Scripts[0].ID != "http-title" {
		t.Errorf("port 80 scripts: %+v", p80.Scripts)
	}
	if h.OS == nil || len(h.OS.Matches) != 1 || h.OS.Matches[0].Accuracy != 98 {
		t.Errorf("os match: %+v", h.OS)
	}
	if progress == nil || progress.Percent < 42 {
		t.Errorf("task progress: %+v", progress)
	}
	if gotStats == nil || gotStats.Hosts.Up != 1 || gotStats.Finished.Elapsed < 3 {
		t.Errorf("run stats: %+v", gotStats)
	}
}

// Truncated XML (mid-host) should still capture what was read and return nil
// error or a recoverable one — we never want a panic.
func TestParseStream_Truncated(t *testing.T) {
	truncated := sampleXML[:len(sampleXML)/2]
	called := false
	_, err := ParseStream(strings.NewReader(truncated), Callbacks{
		OnRunStart: func(string, string, string, int64) { called = true },
	})
	if !called {
		t.Error("RunStart should have fired before truncation")
	}
	// Decoder may report an unexpected-EOF; that is fine, just non-nil is acceptable.
	_ = err
}
