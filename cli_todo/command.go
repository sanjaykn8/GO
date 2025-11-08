package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type CmdFlags struct {
	Add        string
	Delete     int
	Toggle     int
	PrintTable bool
	Edit       string
	Help       bool
}

func NewCmdFlags(todos *Todos) *CmdFlags {
	cf := &CmdFlags{}
	flag.StringVar(&cf.Add, "add", "", "Add a new todo -add \"title")
	flag.StringVar(&cf.Edit, "edit", "", "Edit a todo: -edit index:new_title")
	flag.IntVar(&cf.Delete, "del", -1, "Delete a todo by index: -del N")
	flag.IntVar(&cf.Toggle, "toggle", -1, "Toggle completion by index: -toggle N")
	flag.BoolVar(&cf.PrintTable, "list", false, "List todos")
	flag.BoolVar(&cf.Help, "help", false, "Show help")
	flag.Parse()
	return cf
}

func (cf *CmdFlags) Execute(todos *Todos) {
	switch {
	case cf.Help:
		usage()

	case cf.Add != "":
		todos.Add(cf.Add)

	case cf.Delete != -1:
		if err := todos.Delete(cf.Delete); err != nil {
			fmt.Printf("delete: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("deleted")

	case cf.Edit != "":
		parts := strings.SplitN(cf.Edit, ":", 2)
		if len(parts) != 2 {
			fmt.Println("Error: invalid format for edit. Expecting index:title")
			os.Exit(2)
		}

		idx, err := strconv.Atoi(parts[0])
		if err != nil {
			fmt.Println("Invalid index")
			os.Exit(2)
		}

		if err := todos.Edit(idx, parts[1]); err != nil {
			fmt.Println(err)
			os.Exit(2)
		}

	case cf.Toggle != -1:
		if err := todos.Toggle(cf.Toggle); err != nil {
			fmt.Println(err)
			os.Exit(2)
		}

	case cf.PrintTable:
		todos.PrintTable()

	default:
		usage()
	}
}

func usage() {
	fmt.Println(`Usage:
  -add "title"          Add a new todo
  -list                 List todos
  -toggle N             Toggle todo N
  -del N                Delete todo N
  -edit index:new_title Edit todo at index
  -help                 Show this help`)
}
