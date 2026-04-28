package nmap

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"os/exec"
	"sync"
	"time"
)

func newScanID() string {
	var b [12]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// EventKind tags the type of message streamed from a running scan.
type EventKind string

const (
	EventStart        EventKind = "start"
	EventStderr       EventKind = "stderr"        // raw nmap stderr line (warnings, ETA text)
	EventScanInfo     EventKind = "scaninfo"      // top-level <scaninfo>
	EventHost         EventKind = "host"          // a fully-decoded <host> element
	EventTaskProgress EventKind = "taskprogress"  // periodic --stats-every progress
	EventRunStats     EventKind = "runstats"      // final <runstats> summary
	EventDone         EventKind = "done"
	EventError        EventKind = "error"
)

// Event is one message in a scan's stream. Most kinds carry exactly one of
// the typed payload fields; consumers should switch on Kind.
type Event struct {
	Kind     EventKind     `json:"kind"`
	When     time.Time     `json:"when"`
	Line     string        `json:"line,omitempty"`     // stderr / error
	Code     int           `json:"code,omitempty"`     // done
	Err      string        `json:"err,omitempty"`      // error / done
	ScanInfo *ScanInfo     `json:"scaninfo,omitempty"`
	Host     *Host         `json:"host,omitempty"`
	Progress *TaskProgress `json:"progress,omitempty"`
	RunStats *RunStats     `json:"runstats,omitempty"`
}

// Result is the consolidated record persisted on scan completion.
type Result struct {
	ID       string    `json:"id"`
	Argv     []string  `json:"argv"`
	Display  string    `json:"display"`
	Targets  []string  `json:"targets,omitempty"`
	FlagIDs  []string  `json:"flag_ids,omitempty"`
	Started  time.Time `json:"started"`
	Ended    time.Time `json:"ended"`
	ExitCode int       `json:"exit_code"`
	Run      *Run      `json:"run,omitempty"`     // structured nmap output (nil if scan failed before parse)
	RawXML   []byte    `json:"raw_xml,omitempty"` // full XML bytes for export
	Error    string    `json:"error,omitempty"`
}

// Hook is invoked by the runner exactly once when a scan finishes (success
// or failure). Used to persist Results to history.
type Hook func(Result)

// Scan represents a running or finished scan.
type Scan struct {
	ID      string
	Built   Built
	Started time.Time

	mu          sync.Mutex
	cmd         *exec.Cmd
	cancel      context.CancelFunc
	subscribers []chan Event
	finished    bool
	exitCode    int

	// Captured during the run; available after wait() finishes.
	rawXML  []byte
	runDone *Run
	parseErr error

	finishedAt time.Time
}

// Runner manages concurrent scans by ID and notifies a Hook on completion.
type Runner struct {
	mu    sync.Mutex
	scans map[string]*Scan
	hook  Hook
}

// NewRunner returns a Runner with an empty scan registry.
func NewRunner() *Runner {
	return &Runner{scans: map[string]*Scan{}}
}

// SetHook registers a single completion hook (invoked from a goroutine).
// Pass nil to clear.
func (r *Runner) SetHook(h Hook) {
	r.mu.Lock()
	r.hook = h
	r.mu.Unlock()
}

// Start spawns nmap with the given Built command. Build must already have
// validated privileges and version constraints.
func (r *Runner) Start(parent context.Context, b Built) (*Scan, error) {
	if len(b.Argv) == 0 {
		return nil, errors.New("empty argv")
	}
	ctx, cancel := context.WithCancel(parent)
	cmd := exec.CommandContext(ctx, b.Argv[0], b.Argv[1:]...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, err
	}
	scan := &Scan{
		ID:      newScanID(),
		Built:   b,
		Started: time.Now(),
		cmd:     cmd,
		cancel:  cancel,
	}
	r.mu.Lock()
	r.scans[scan.ID] = scan
	r.mu.Unlock()

	scan.publish(Event{Kind: EventStart, When: time.Now()})

	stdoutDone := make(chan struct{})
	go scan.parseStdout(stdout, stdoutDone)
	go scan.pumpStderr(stderr)
	go r.wait(scan, stdoutDone)
	return scan, nil
}

// Get returns a registered scan by ID.
func (r *Runner) Get(id string) (*Scan, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.scans[id]
	return s, ok
}

// Stop cancels a running scan via context cancellation; the spawned process
// receives SIGKILL via exec.CommandContext. Idempotent.
func (s *Scan) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.finished || s.cancel == nil {
		return
	}
	s.cancel()
}

// Subscribe returns a channel that receives all subsequent Events for this
// scan plus a final Done event. The channel is closed after Done.
func (s *Scan) Subscribe() <-chan Event {
	ch := make(chan Event, 256)
	s.mu.Lock()
	if s.finished {
		// Late subscriber: synthesize a Done immediately.
		go func() {
			ch <- Event{Kind: EventDone, When: time.Now(), Code: s.exitCode}
			close(ch)
		}()
		s.mu.Unlock()
		return ch
	}
	s.subscribers = append(s.subscribers, ch)
	s.mu.Unlock()
	return ch
}

// Result returns the consolidated record. Only meaningful after the scan has
// finished; before that, Ended is the zero time.
func (s *Scan) Result() Result {
	s.mu.Lock()
	defer s.mu.Unlock()
	res := Result{
		ID:       s.ID,
		Argv:     s.Built.Argv,
		Display:  s.Built.Display,
		Targets:  s.Built.Targets,
		FlagIDs:  s.Built.FlagIDs,
		Started:  s.Started,
		Ended:    s.finishedAt,
		ExitCode: s.exitCode,
		Run:      s.runDone,
		RawXML:   s.rawXML,
	}
	if s.parseErr != nil {
		res.Error = s.parseErr.Error()
	}
	return res
}

func (s *Scan) publish(ev Event) {
	s.mu.Lock()
	subs := append([]chan Event(nil), s.subscribers...)
	s.mu.Unlock()
	for _, ch := range subs {
		select {
		case ch <- ev:
		default: // drop on slow subscriber rather than blocking the runner
		}
	}
}

func (s *Scan) parseStdout(r io.Reader, done chan<- struct{}) {
	defer close(done)
	run := &Run{}
	cb := Callbacks{
		OnRunStart: func(scanner, args, version string, start int64) {
			run.Scanner = scanner
			run.Args = args
			run.Version = version
			run.Start = start
		},
		OnScanInfo: func(si ScanInfo) {
			run.ScanInfo = append(run.ScanInfo, si)
			s.publish(Event{Kind: EventScanInfo, When: time.Now(), ScanInfo: &si})
		},
		OnHost: func(h Host) {
			run.Hosts = append(run.Hosts, h)
			hh := h
			s.publish(Event{Kind: EventHost, When: time.Now(), Host: &hh})
		},
		OnTaskProgress: func(p TaskProgress) {
			pp := p
			s.publish(Event{Kind: EventTaskProgress, When: time.Now(), Progress: &pp})
		},
		OnRunStats: func(rs RunStats) {
			run.RunStats = &rs
			rsCopy := rs
			s.publish(Event{Kind: EventRunStats, When: time.Now(), RunStats: &rsCopy})
		},
	}
	raw, err := ParseStream(r, cb)
	s.mu.Lock()
	s.rawXML = raw
	s.runDone = run
	s.parseErr = err
	s.mu.Unlock()
}

func (s *Scan) pumpStderr(r io.Reader) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		s.publish(Event{Kind: EventStderr, Line: sc.Text(), When: time.Now()})
	}
}

func (r *Runner) wait(s *Scan, stdoutDone <-chan struct{}) {
	cmdErr := s.cmd.Wait()
	<-stdoutDone // ensure the parser has fully consumed stdout

	now := time.Now()
	s.mu.Lock()
	s.finished = true
	s.finishedAt = now
	if s.cmd.ProcessState != nil {
		s.exitCode = s.cmd.ProcessState.ExitCode()
	}
	subs := s.subscribers
	s.subscribers = nil
	s.mu.Unlock()

	done := Event{Kind: EventDone, When: now, Code: s.exitCode}
	if cmdErr != nil {
		done.Err = cmdErr.Error()
	}
	for _, ch := range subs {
		select {
		case ch <- done:
		default:
		}
		close(ch)
	}

	r.mu.Lock()
	hook := r.hook
	r.mu.Unlock()
	if hook != nil {
		hook(s.Result())
	}
}
