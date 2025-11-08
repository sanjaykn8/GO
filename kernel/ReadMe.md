# 🧠 GoLang Mini Projects

A collection of four small but powerful Go projects built while exploring systems, web, and CLI programming.  
Each project focuses on a different aspect of Go — from servers to schedulers.

---

## 🚀 Projects Overview

### 1. URL Shortener (`url_shortener`)
A lightweight URL shortening service built using the **Gin** web framework.

**Concepts:** REST APIs, HTTP handlers, JSON parsing, routing, and persistence.

**Run:**
```bash
cd url_shortener
go run .
````

---

### 2. Simple HTTP Server (`simple_http`)

A minimal HTTP server built using Go’s native `net/http` package.

**Concepts:** Request handling, routing, templates, and concurrency with goroutines.

**Run:**

```bash
cd simple_http
go run .
```

---

### 3. CLI Todo App (`cli_todo`)

A terminal-based to-do list app that supports adding, editing, listing, and removing tasks.

**Concepts:** Command-line flags, file storage, and structured error handling.

**Run:**

```bash
cd cli_todo
go run . -add "Finish report"
go run . -list
```

---

### 4. GoSimOS — Simple Kernel Simulator (`kernel`)

A simulated **mini operating system kernel**, demonstrating how schedulers, processes, and inter-process communication (IPC) work.

**Concepts:** Process scheduling, concurrency, message passing, and virtual file systems.

**Run (example):**

```bash
cd kernel
go run . -demo -procs 32 -min 3 -max 10 -secs 10
```

**Sample Output:**

```
🖥️ GoSimOS — lightweight kernel simulator
Quantum: 100ms | MaxRun: 10s

Process Summary
──────────────────────────────────────────────
PID  Name            Priority  CPU(ms)  Status
 1   proc-01         0         400ms    Completed
 2   proc-02         1         600ms    Completed
...
✅ Simulation finished (elapsed: 9.84s, processes: 32)
```

---

## 🧩 Tech Stack

* **Language:** Go 1.25
* **Libraries:** `gin`, `flag`, `net/http`, `sync/atomic`, `math/rand`
* **Paradigms:** Concurrency, Channels, Modular Design, CLI tools

---

## 🧠 Learning Outcomes

* Build and deploy web servers with Go
* Handle concurrency and race-free state
* Design and test CLI utilities
* Simulate OS-level scheduling and IPC

---

## 📂 Repository Structure

```
GO/
├── url_shortener/
├── simple_http/
├── cli_todo/
└── kernel/
```

## 📜 License

MIT License — feel free to use or modify with credit.

```
