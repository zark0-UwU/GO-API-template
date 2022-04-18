package main

import (
	"ToDoList/src"
	"log"
)

func main() {
	defer log.Println("Shutting down")
	src.Start()
}
