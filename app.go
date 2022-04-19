package main

import (
	"log"
	"template/src"
)

func main() {
	defer log.Println("Shutting down")
	src.Start()
}
