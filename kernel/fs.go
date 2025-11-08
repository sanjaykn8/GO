package main

import (
	"sync"
)

type SimFS struct {
	mu    sync.Mutex
	files map[string]string
}

func NewSimFS() *SimFS {
	return &SimFS{
		files: make(map[string]string),
	}
}

func (f *SimFS) WriteFile(name, content string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.files[name] = content
	return nil
}

func (f *SimFS) ReadFile(name string) (string, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	content, ok := f.files[name]
	return content, ok
}

func (f *SimFS) Dump() map[string]string {
	f.mu.Lock()
	defer f.mu.Unlock()

	dup := make(map[string]string, len(f.files))
	for k, v := range f.files {
		dup[k] = v
	}

	return dup
}
