package main

import (
	"sort"
	"sync"
	"time"
)

type Message struct {
	From    int
	To      int
	Payload string
}

type ProcessStat struct {
	ID        int
	Name      string
	Priority  int
	RunCount  int
	TotalCPU  time.Duration
	Remaining int
}

type Scheduler struct {
	quantum   time.Duration
	mu        sync.Mutex
	ready     []*Process
	procs     map[int]*Process
	mailboxes map[int][]Message
	fs        *SimFS
	running   bool
	stopCh    chan struct{}
	doneCh    chan struct{}
}

func NewScheduler(quantum time.Duration) *Scheduler {
	return &Scheduler{
		quantum:   quantum,
		ready:     []*Process{},
		procs:     make(map[int]*Process),
		mailboxes: make(map[int][]Message),
		fs:        NewSimFS(),
		stopCh:    make(chan struct{}),
		doneCh:    make(chan struct{}),
	}
}

func (s *Scheduler) Spawn(spec *ProcessSpec) int {
	p := NewProcess(spec)
	s.mu.Lock()
	s.ready = append(s.ready, p)
	s.procs[p.ID] = p
	s.mailboxes[p.ID] = []Message{}
	s.mu.Unlock()
	return p.ID
}

func (s *Scheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}

	s.running = true
	s.mu.Unlock()

	go s.loop()
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}

	close(s.stopCh)
	s.mu.Unlock()
	<-s.doneCh
}

func (s *Scheduler) SendMessage(from, to int, msg Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.mailboxes[to]; ok {
		s.mailboxes[to] = append(s.mailboxes[to], msg)
	}
}

func (s *Scheduler) loop() {
	defer close(s.doneCh)

	for {
		select {
		case <-s.stopCh:
			return
		default:
		}

		s.mu.Lock()
		if len(s.ready) == 0 {
			s.mu.Unlock()
			time.Sleep(50 * time.Millisecond)
			continue
		}

		sort.SliceStable(s.ready, func(i, j int) bool {
			return s.ready[i].Priority < s.ready[j].Priority
		})
		p := s.ready[0]

		if len(s.ready) == 1 {
			s.ready = []*Process{}
		} else {
			s.ready = append(s.ready[:0], s.ready[1:]...)
		}
		s.mu.Unlock()

		finished := p.Run(s.quantum, s.fs, s)

		if !finished {
			s.mu.Lock()
			s.ready = append(s.ready, p)
			s.mu.Unlock()
		} else {
			// mark complete, keep in procs for stats but not in ready queue
		}
	}
}

func (s *Scheduler) Stats() []ProcessStat {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]ProcessStat, 0, len(s.procs))

	for _, p := range s.procs {
		remaining := int(p.WorkUnits)
		out = append(out, ProcessStat{
			ID:        p.ID,
			Name:      p.Name,
			Priority:  p.Priority,
			RunCount:  p.RunCount,
			TotalCPU:  p.TotalCPU,
			Remaining: remaining,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})

	return out
}

func (s *Scheduler) DumpMailboxes() map[int][]Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	dup := make(map[int][]Message, len(s.mailboxes))

	for k, v := range s.mailboxes {
		dup[k] = append([]Message(nil), v...)
	}

	return dup
}

func (s *Scheduler) DumpFS() map[string]string {
	return s.fs.Dump()
}
