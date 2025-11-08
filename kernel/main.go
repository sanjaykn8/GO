package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

const (
	ansiReset   = "\033[0m"
	ansiBold    = "\033[1m"
	ansiRed     = "\033[31m"
	ansiGreen   = "\033[32m"
	ansiYellow  = "\033[33m"
	ansiCyan    = "\033[36m"
	ansiMagenta = "\033[35m"
)

func main() {
	var quantumMs int
	var demo bool
	var runSecs int

	var procCount int
	var minUnits int
	var maxUnits int
	var randomize bool
	var seedVal int64

	flag.IntVar(&procCount, "procs", 16, "number of processes to spawn")
	flag.IntVar(&minUnits, "min", 2, "min work units per process")
	flag.IntVar(&maxUnits, "max", 12, "max work units per process")
	flag.BoolVar(&randomize, "random", true, "randomize names/workloads")
	flag.Int64Var(&seedVal, "seed", time.Now().UnixNano(), "rng seed (0 = deterministic)")

	flag.IntVar(&quantumMs, "quantum", 100, "CPU quantum in ms")
	flag.BoolVar(&demo, "demo", true, "run test scenario")
	flag.IntVar(&runSecs, "secs", 6, "Max. seconds")
	flag.Parse()

	if demo {
		runDemo(time.Duration(quantumMs)*time.Millisecond,
			time.Duration(runSecs)*time.Second,
			procCount, minUnits, maxUnits, randomize, seedVal)
		return
	}
	fmt.Println("No mode selected. Use -demo or own process")
}

func runDemo(quantum time.Duration, maxRun time.Duration, procCount, minUnits, maxUnits int, randomize bool, seedVal int64) {
	start := time.Now()
	log.Printf("OS starting: Quantum = %v MaxRun = %v\n", quantum, maxRun)

	s := NewScheduler(quantum)

	rng := rand.New(rand.NewSource(seedVal))
	names := []string{"worker", "io", "net", "db", "logger", "ipc", "fs", "cache"}

	for i := 0; i < procCount; i++ {
		name := fmt.Sprintf("proc-%02d", i+1)
		behavior := BehaviorCompute
		prio := rng.Intn(3)

		if randomize {
			r := rng.Intn(100)
			if r < 10 {
				behavior = BehaviorFSWriter
			} else if r < 30 {
				behavior = BehaviorIPCSender
			}
		} else {
			if i%7 == 0 {
				behavior = BehaviorFSWriter
			} else if i%5 == 0 {
				behavior = BehaviorIPCSender
			}
		}

		wu := minUnits
		if maxUnits > minUnits {
			wu = minUnits + rng.Intn(maxUnits-minUnits+1)
		}
		s.Spawn(&ProcessSpec{
			Name:      fmt.Sprintf("%s-%s", names[i%len(names)], name),
			Priority:  prio,
			WorkUnits: wu,
			Behavior:  behavior,
		})
	}

	clearScreenIfTTY()
	printBanner()
	fmt.Printf("%s🖥️  %sGoSimOS%s — lightweight kernel simulator\n", ansiBold, ansiCyan, ansiReset)
	fmt.Printf("%sQuantum:%s %s | %sMaxRun:%s %s\n\n", ansiBold, ansiReset, quantum, ansiBold, ansiReset, maxRun)

	s.Start()
	time.Sleep(maxRun)
	s.Stop()

	elapsed := time.Since(start)

	fmt.Println()
	printDivider()
	fmt.Printf("%sProcess Summary%s\n", ansiBold, ansiReset)
	printDivider()
	printProcessTable(s.Stats())
	fmt.Println()

	printDivider()
	fmt.Printf("%sMailboxes%s\n", ansiBold, ansiReset)
	printDivider()
	printMailboxes(s.DumpMailboxes())
	fmt.Println()

	printDivider()
	fmt.Printf("%sVirtual FS%s\n", ansiBold, ansiReset)
	printDivider()
	printFS(s.DumpFS())
	fmt.Println()

	printDivider()
	fmt.Printf("%s✅ Simulation finished%s  (elapsed: %s, processes: %d)\n",
		ansiGreen, ansiReset, elapsed.Round(time.Millisecond), len(s.Stats()))
	fmt.Println()
	_ = os.WriteFile("gosimos_summary.txt", []byte(plainSummary(s, elapsed)), 0644)
	fmt.Printf("%sSaved text summary to gosimos_summary.txt%s\n", ansiYellow, ansiReset)
}

func clearScreenIfTTY() {
	fmt.Print("\033[2J\033[H")
}

func printBanner() {
	b := `
  ____        ____  _ 
 / ___| ___  / ___|(_)__    __
| |  _ / _ \ \___ \| ||\\  //||
| |_| | |_| | ___) | || \\// ||
 \____|\___/ |____/|_||	     ||
`
	fmt.Printf("%s%s%s\n", ansiMagenta, b, ansiReset)
}

func printDivider() {
	fmt.Println(strings.Repeat("─", 60))
}

func printProcessTable(stats []ProcessStat) {
	fmt.Printf("%s%3s  %-16s  %-7s  %-8s  %s%s\n", ansiBold, "PID", "Name", "Priority", "CPU", "Status", ansiReset)
	for _, st := range stats {
		status := fmt.Sprintf("%sCompleted%s", ansiGreen, ansiReset)
		if st.Remaining > 0 {
			status = fmt.Sprintf("%sRunning (%d left)%s", ansiYellow, st.Remaining, ansiReset)
		}
		priColor := ansiCyan
		if st.Priority == 0 {
			priColor = ansiRed
		}
		fmt.Printf(" %3d  %-16s  %s%-7d%s  %6s  %s\n",
			st.ID,
			truncate(st.Name, 16),
			priColor, st.Priority, ansiReset,
			st.TotalCPU.Round(time.Millisecond),
			status)
	}
}

func printMailboxes(m map[int][]Message) {
	if len(m) == 0 {
		fmt.Println(" (none)")
		return
	}
	for pid, msgs := range m {
		if len(msgs) == 0 {
			fmt.Printf(" PID %2d  ←  %s—%s\n", pid, ansiYellow, ansiReset)
			continue
		}
		preview := msgsPreview(msgs)
		fmt.Printf(" PID %2d  ←  %s%d message(s)%s  preview: %s\n",
			pid, ansiGreen, len(msgs), ansiReset, preview)
	}
}

func msgsPreview(msgs []Message) string {
	n := len(msgs)
	if n == 0 {
		return "—"
	}
	limit := 3
	if n < limit {
		limit = n
	}
	parts := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		parts = append(parts, fmt.Sprintf("[%d→%d %s]", msgs[i].From, msgs[i].To, truncate(msgs[i].Payload, 20)))
	}
	if n > limit {
		parts = append(parts, fmt.Sprintf("…(+%d)", n-limit))
	}
	return strings.Join(parts, " ")
}

func printFS(fs map[string]string) {
	if len(fs) == 0 {
		fmt.Println(" (empty)")
		return
	}
	for name, content := range fs {
		fmt.Printf(" %s%s%s  →  %q\n", ansiBold, name, ansiReset, truncate(content, 60))
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

func plainSummary(s *Scheduler, elapsed time.Duration) string {
	sb := &strings.Builder{}
	sb.WriteString("GoSimOS — simulation summary\n")
	sb.WriteString(strings.Repeat("-", 40) + "\n")
	sb.WriteString(fmt.Sprintf("Quantum: %v  Elapsed: %v\n\n", s.quantum, elapsed))
	sb.WriteString("Processes:\n")
	for _, st := range s.Stats() {
		sb.WriteString(fmt.Sprintf(" PID=%d name=%s prio=%d cpu=%v remaining=%d\n",
			st.ID, st.Name, st.Priority, st.TotalCPU.Round(time.Millisecond), st.Remaining))
	}
	sb.WriteString("\nMailboxes:\n")
	for k, v := range s.DumpMailboxes() {
		sb.WriteString(fmt.Sprintf(" PID=%d messages=%d\n", k, len(v)))
	}
	sb.WriteString("\nFiles:\n")
	for k, v := range s.DumpFS() {
		sb.WriteString(fmt.Sprintf(" %s -> %q\n", k, v))
	}
	return sb.String()
}

//Moderate: go run . -demo -procs 32 -min 3 -max 12 -secs 10
//Heavy: go run . -demo -procs 128 -min 5 -max 30 -secs 40
//Deterministic: go run . -demo -procs 64 -min 3 -max 10 -seed 12345 -random=false -secs 12
