package nmap

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
)

// Run is the top-level <nmaprun> element. Only fields we surface are mapped.
type Run struct {
	Scanner   string     `xml:"scanner,attr"  json:"scanner"`
	Args      string     `xml:"args,attr"     json:"args"`
	Start     int64      `xml:"start,attr"    json:"start"`
	StartStr  string     `xml:"startstr,attr" json:"startstr,omitempty"`
	Version   string     `xml:"version,attr"  json:"version"`
	XMLVer    string     `xml:"xmloutputversion,attr" json:"xmloutputversion,omitempty"`
	ScanInfo  []ScanInfo `xml:"scaninfo"      json:"scaninfo,omitempty"`
	Hosts     []Host     `xml:"host"          json:"hosts"`
	RunStats  *RunStats  `xml:"runstats"      json:"runstats,omitempty"`
}

type ScanInfo struct {
	Type        string `xml:"type,attr"        json:"type"`
	Protocol    string `xml:"protocol,attr"    json:"protocol"`
	NumServices int    `xml:"numservices,attr" json:"numservices"`
	Services    string `xml:"services,attr"    json:"services,omitempty"`
}

type Host struct {
	StartTime int64      `xml:"starttime,attr" json:"starttime,omitempty"`
	EndTime   int64      `xml:"endtime,attr"   json:"endtime,omitempty"`
	Status    Status     `xml:"status"         json:"status"`
	Addresses []Address  `xml:"address"        json:"addresses"`
	Hostnames Hostnames  `xml:"hostnames"      json:"hostnames"`
	Ports     *Ports     `xml:"ports"          json:"ports,omitempty"`
	OS        *OS        `xml:"os"             json:"os,omitempty"`
	Times     *HostTimes `xml:"times"          json:"times,omitempty"`
}

type Status struct {
	State     string  `xml:"state,attr"     json:"state"`
	Reason    string  `xml:"reason,attr"    json:"reason,omitempty"`
	ReasonTTL int     `xml:"reason_ttl,attr" json:"reason_ttl,omitempty"`
}

type Address struct {
	Addr     string `xml:"addr,attr"     json:"addr"`
	AddrType string `xml:"addrtype,attr" json:"addrtype"`
	Vendor   string `xml:"vendor,attr"   json:"vendor,omitempty"`
}

type Hostnames struct {
	Hostnames []Hostname `xml:"hostname" json:"hostname,omitempty"`
}

type Hostname struct {
	Name string `xml:"name,attr" json:"name"`
	Type string `xml:"type,attr" json:"type"`
}

type Ports struct {
	ExtraPorts []ExtraPorts `xml:"extraports" json:"extraports,omitempty"`
	Ports      []Port       `xml:"port"       json:"ports,omitempty"`
}

type ExtraPorts struct {
	State string `xml:"state,attr" json:"state"`
	Count int    `xml:"count,attr" json:"count"`
}

type Port struct {
	Protocol string    `xml:"protocol,attr" json:"protocol"`
	PortID   int       `xml:"portid,attr"   json:"portid"`
	State    PortState `xml:"state"         json:"state"`
	Service  *Service  `xml:"service"       json:"service,omitempty"`
	Scripts  []Script  `xml:"script"        json:"scripts,omitempty"`
}

type PortState struct {
	State     string `xml:"state,attr"      json:"state"`
	Reason    string `xml:"reason,attr"     json:"reason,omitempty"`
	ReasonTTL int    `xml:"reason_ttl,attr" json:"reason_ttl,omitempty"`
}

type Service struct {
	Name      string `xml:"name,attr"       json:"name"`
	Product   string `xml:"product,attr"    json:"product,omitempty"`
	Version   string `xml:"version,attr"    json:"version,omitempty"`
	ExtraInfo string `xml:"extrainfo,attr"  json:"extrainfo,omitempty"`
	Method    string `xml:"method,attr"     json:"method,omitempty"`
	Conf      int    `xml:"conf,attr"       json:"conf,omitempty"`
	OSType    string `xml:"ostype,attr"     json:"ostype,omitempty"`
	CPE       []string `xml:"cpe"           json:"cpe,omitempty"`
}

type Script struct {
	ID     string `xml:"id,attr"     json:"id"`
	Output string `xml:"output,attr" json:"output"`
}

type OS struct {
	Matches []OSMatch `xml:"osmatch" json:"matches,omitempty"`
}

type OSMatch struct {
	Name     string    `xml:"name,attr"     json:"name"`
	Accuracy int       `xml:"accuracy,attr" json:"accuracy"`
	Classes  []OSClass `xml:"osclass"       json:"classes,omitempty"`
}

type OSClass struct {
	Type     string `xml:"type,attr"     json:"type,omitempty"`
	Vendor   string `xml:"vendor,attr"   json:"vendor,omitempty"`
	OSFamily string `xml:"osfamily,attr" json:"osfamily,omitempty"`
	OSGen    string `xml:"osgen,attr"    json:"osgen,omitempty"`
	Accuracy int    `xml:"accuracy,attr" json:"accuracy,omitempty"`
}

type HostTimes struct {
	SRTT   int `xml:"srtt,attr"   json:"srtt,omitempty"`
	RTTVar int `xml:"rttvar,attr" json:"rttvar,omitempty"`
	TO     int `xml:"to,attr"     json:"to,omitempty"`
}

type RunStats struct {
	Finished Finished `xml:"finished" json:"finished"`
	Hosts    HostsAgg `xml:"hosts"    json:"hosts"`
}

type Finished struct {
	Time    int64   `xml:"time,attr"    json:"time"`
	TimeStr string  `xml:"timestr,attr" json:"timestr,omitempty"`
	Elapsed float64 `xml:"elapsed,attr" json:"elapsed"`
	Summary string  `xml:"summary,attr" json:"summary,omitempty"`
	Exit    string  `xml:"exit,attr"    json:"exit,omitempty"`
}

type HostsAgg struct {
	Up    int `xml:"up,attr"    json:"up"`
	Down  int `xml:"down,attr"  json:"down"`
	Total int `xml:"total,attr" json:"total"`
}

// TaskProgress is emitted periodically when --stats-every is set.
type TaskProgress struct {
	Task      string  `xml:"task,attr"      json:"task,omitempty"`
	Time      int64   `xml:"time,attr"      json:"time,omitempty"`
	Percent   float64 `xml:"percent,attr"   json:"percent"`
	Remaining int     `xml:"remaining,attr" json:"remaining,omitempty"`
	ETC       int64   `xml:"etc,attr"       json:"etc,omitempty"`
}

// Callbacks is the streaming-parser callback set. Any field may be nil.
type Callbacks struct {
	OnRunStart    func(scanner, args, version string, start int64)
	OnScanInfo    func(ScanInfo)
	OnHost        func(Host)
	OnTaskProgress func(TaskProgress)
	OnRunStats    func(RunStats)
}

// ParseStream consumes nmap XML from r, invokes callbacks as elements close,
// and returns the full raw bytes that were read (suitable for storage).
//
// EOF mid-stream (e.g. when nmap is killed before </nmaprun>) is reported
// via the returned error but the caller still receives whatever bytes and
// callbacks were emitted up to that point.
func ParseStream(r io.Reader, cb Callbacks) ([]byte, error) {
	var buf bytes.Buffer
	tee := io.TeeReader(r, &buf)
	dec := xml.NewDecoder(tee)
	dec.Strict = false
	for {
		tok, err := dec.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return buf.Bytes(), nil
			}
			return buf.Bytes(), err
		}
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch start.Name.Local {
		case "nmaprun":
			if cb.OnRunStart != nil {
				cb.OnRunStart(
					attr(start, "scanner"),
					attr(start, "args"),
					attr(start, "version"),
					atoi64(attr(start, "start")),
				)
			}
		case "scaninfo":
			var s ScanInfo
			if err := dec.DecodeElement(&s, &start); err == nil && cb.OnScanInfo != nil {
				cb.OnScanInfo(s)
			}
		case "host":
			var h Host
			if err := dec.DecodeElement(&h, &start); err != nil {
				return buf.Bytes(), fmt.Errorf("decoding host: %w", err)
			}
			if cb.OnHost != nil {
				cb.OnHost(h)
			}
		case "taskprogress":
			var p TaskProgress
			if err := dec.DecodeElement(&p, &start); err == nil && cb.OnTaskProgress != nil {
				cb.OnTaskProgress(p)
			}
		case "runstats":
			var rs RunStats
			if err := dec.DecodeElement(&rs, &start); err == nil && cb.OnRunStats != nil {
				cb.OnRunStats(rs)
			}
		}
	}
}

// PrimaryAddress returns the first non-MAC address attached to a host.
func (h Host) PrimaryAddress() string {
	for _, a := range h.Addresses {
		if a.AddrType == "ipv4" || a.AddrType == "ipv6" {
			return a.Addr
		}
	}
	for _, a := range h.Addresses {
		return a.Addr
	}
	return ""
}

func attr(start xml.StartElement, name string) string {
	for _, a := range start.Attr {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

func atoi64(s string) int64 {
	var n int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return n
		}
		n = n*10 + int64(c-'0')
	}
	return n
}
