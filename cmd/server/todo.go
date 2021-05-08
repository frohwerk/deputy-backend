package main

import "fmt"

type TodoList []string

var todos = TodoList{
	"TODO: cmd/server/apps/get.go - Handle request without env query parameter",
}

func (l TodoList) print() {
	for _, todo := range l {
		fmt.Println(todo)
	}
}
