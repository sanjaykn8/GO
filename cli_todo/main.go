package main

import "log"

const todosfile = "todos.json"

func main() {
	todos := Todos{}
	st := NewStorage[Todos](todosfile)

	_ = st.Load(&todos)
	flags := NewCmdFlags(&todos)
	flags.Execute(&todos)

	if err := st.save(&todos); err != nil {
		log.Fatalf("failed to save todos: %v", err)
	}
}
