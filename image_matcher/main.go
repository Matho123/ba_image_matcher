package main

import (
	"image-matcher/image_matcher/testing"
	"log"
	"os"
)

func main() {
	command := os.Args[1]
	arguments := os.Args[2:]

	commandFunction := testing.CommandMapping[command]

	if commandFunction == nil {
		log.Fatal("Not a valid command!")
	}
	commandFunction(arguments)
}
