package main

import (
	"testing"
	"time"
)

func TestPriorityHighFirst(t *testing.T) {
	s := NewScheduler(100 * time.Millisecond)
	defer s.Stop()

	p1 := &ProcessSpec{
		Name:      "high",
		Priority:  0,
		WorkUnits: 3,
		Behavior:  BehaviorCompute,
	}
	p2 := &ProcessSpec{
		Name:      "low",
		Priority:  2,
		WorkUnits: 6,
		Behavior:  BehaviorCompute,
	}

	s.Spawn(p2)
	s.Spawn(p1)

	s.Start()
	time.Sleep(1200 * time.Millisecond)
	s.Stop()

	stats := s.Stats()
	gotHighRemaining := -1
	gotLowRemaining := -1

	for _, st := range stats {
		if st.Name == "high" {
			gotHighRemaining = st.Remaining
		} else if st.Name == "low" {
			gotLowRemaining = st.Remaining
		}
	}

	if gotHighRemaining > gotLowRemaining {
		t.Errorf("High priority process has more remaining work units than low priority process: high=%d low=%d", gotHighRemaining, gotLowRemaining)
	}
}
