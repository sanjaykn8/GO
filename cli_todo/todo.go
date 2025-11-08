package main

import (
	"errors"
	"fmt"
	"os"
	"text/tabwriter"
	"time"
)

type Todo struct {
	Title       string     `json:"title"`
	Completed   bool       `json:"completed"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type Todos []Todo

func (todos *Todos) Add(title string) {
	t := Todo{
		Title:       title,
		Completed:   false,
		CreatedAt:   time.Now(),
		CompletedAt: nil,
	}

	*todos = append(*todos, t)
}

func (todos *Todos) ValidateIndex(index int) error {
	if index < 0 || index >= len(*todos) {
		return errors.New("invalid index")
	}

	return nil
}

func (todos *Todos) Delete(index int) error {
	if err := todos.ValidateIndex(index); err != nil {
		return err
	}

	t := *todos
	*todos = append(t[:index], t[index+1:]...)

	return nil
}

func (todos *Todos) Toggle(index int) error {
	if err := todos.ValidateIndex(index); err != nil {
		return err
	}

	t := *todos

	if !t[index].Completed {
		t[index].Completed = false
		t[index].CompletedAt = nil
	} else {
		t[index].Completed = true
		now := time.Now()
		t[index].CompletedAt = &now
	}

	t[index].Completed = !t[index].Completed

	return nil
}

func (todos *Todos) Edit(index int, newTitle string) error {
	if err := todos.ValidateIndex(index); err != nil {
		return err
	}

	t := *todos
	t[index].Title = newTitle

	return nil
}

func (todos *Todos) PrintTable() {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintf(w, "#\tTitle\tCompleted\tCreated At\tCompleted At\n")
	for i, t := range *todos {
		comp := "❌"
		compAt := ""
		if t.Completed {
			comp = "✅"
			if t.CompletedAt != nil {
				compAt = t.CompletedAt.Format(time.RFC1123)
			}
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n",
			i,
			t.Title,
			comp,
			t.CreatedAt.Format(time.RFC1123),
			compAt,
		)
	}
	_ = w.Flush()
}
