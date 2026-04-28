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
	EventStdout EventKind = "stdout" // raw nmap stdout line (XML; consumers can reparse)
	EventStderr EventKind = "stderr" // raw nmap stderr line (status, warnings)
	EventStart  EventKind = "start"
	EventDone   EventKind = "done"
	EventError  EventKind = "error"
)

// Event is one message in a scan's stream.
type Event struct {
	Kind EventKind `json:"kind"`
	Line string    `json:"line,omitempty"`
	When time.Time `json:"when"`
	Code int       `json:"code,omitempty"` // exit code on Done; 0 otherwise
	Err  string    `json:"err,omitempty"`
}

// Scan represents a running or finished scan. Subscribers receive a stream of
// Events on a per-call channel.
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
}

// Runner manages concurrent scans by ID.
type Runner struct {
	mu    sync.Mutex
	scans map[string]*Scan
}

// NewRunner returns a Runner with an empty scan registry.
func NewRunner() *Runner {
	return &Runner{scans: map[string]*Scan{}}
}

// Start spawns nmap with the given Built command and returns a Scan whose
// events can be subscribed to. The caller is responsible for ensuring Build
// has already validated privileges.
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
	go scan.pump(stdout, EventStdout)
	go scan.pump(stderr, EventStderr)
	go scan.wait()
	return scan, nil
}

// Get returns a registered scan by ID.
func (r *Runner) Get(id string) (*Scan, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.scans[id]
	return s, ok
}

// Stop sends SIGINT to a running scan; it will be force-killed when the parent
// context is cancelled if it doesn't exit gracefully.
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
	ch := make(chan Event, 64)
	s.mu.Lock()
	if s.finished {
		// Late subscriber: send a synthetic done immediately.
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

func (s *Scan) pump(r io.Reader, kind EventKind) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		s.publish(Event{Kind: kind, Line: sc.Text(), When: time.Now()})
	}
}

func (s *Scan) wait() {
	err := s.cmd.Wait()
	s.mu.Lock()
	s.finished = true
	if s.cmd.ProcessState != nil {
		s.exitCode = s.cmd.ProcessState.ExitCode()
	}
	subs := s.subscribers
	s.subscribers = nil
	s.mu.Unlock()

	done := Event{Kind: EventDone, When: time.Now(), Code: s.exitCode}
	if err != nil {
		done.Err = err.Error()
	}
	for _, ch := range subs {
		select {
		case ch <- done:
		default:
		}
		close(ch)
	}
}
