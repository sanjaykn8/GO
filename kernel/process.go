package main

import (
	"fmt"
	"sync/atomic"
	"time"
)

type Behavior int

const (
	BehaviorCompute Behavior = iota
	BehaviorIPCSender
	BehaviorFSWriter
)

type ProcessSpec struct {
	Name      string
	Priority  int //0 -> High
	WorkUnits int
	Behavior  Behavior
}

var pidCounter int32 = 0

type Process struct {
	ID        int
	Name      string
	Priority  int
	WorkUnits int32
	Behavior  Behavior
	RunCount  int
	TotalCPU  time.Duration
	Mailbox   []string
	mailMutex chan struct{}
	fsWrites  []string
	createdAt time.Time
}

func NewProcess(spec *ProcessSpec) *Process {
	id := int(atomic.AddInt32(&pidCounter, 1))
	p := &Process{
		ID:        id,
		Name:      spec.Name,
		Priority:  spec.Priority,
		WorkUnits: int32(spec.WorkUnits),
		Behavior:  spec.Behavior,
		mailMutex: make(chan struct{}, 1),
		createdAt: time.Now(),
	}

	p.mailMutex <- struct{}{}
	return p
}

func (p *Process) Run(quantum time.Duration, fs *SimFS, sched *Scheduler) (finished bool) {
	unit := 100 * time.Millisecond

	maxUnits := int(quantum / unit)
	if maxUnits < 1 {
		maxUnits = 1
	}

	remaining := int(atomic.LoadInt32(&p.WorkUnits))
	if remaining < 0 {
		return true
	}

	toRun := remaining
	if toRun > maxUnits {
		toRun = maxUnits
	}

	start := time.Now()
	time.Sleep(time.Duration(toRun) * unit)
	elapsed := time.Since(start)
	p.TotalCPU += elapsed
	p.RunCount++

	atomic.AddInt32(&p.WorkUnits, -int32(toRun))

	switch p.Behavior {
	case BehaviorIPCSender:
		if target := 1; target != p.ID {
			sched.SendMessage(p.ID, target, Message{
				From:    p.ID,
				To:      target,
				Payload: fmt.Sprintf("MSG from %s at %v", p.Name, time.Since(p.createdAt)),
			})
		}
	case BehaviorFSWriter:
		name := fmt.Sprintf("file_%d.txt", p.ID)
		content := fmt.Sprintf("Data written by %s at %s", p.Name, time.Now().Format(time.RFC3339))
		_ = fs.WriteFile(name, content)
		p.fsWrites = append(p.fsWrites, name)
	}

	remainingAfter := int(atomic.LoadInt32(&p.WorkUnits))
	return remainingAfter <= 0
}
