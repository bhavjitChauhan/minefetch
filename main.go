package main

import (
	"log"
)

var version = "dev"

func main() {
	log.SetFlags(0)

	err := parseArgs()
	if err != nil {
		log.Fatalf("Failed to parse arguments: %v\nSee minefetch --help\n", err)
	}

	results := getResults()

	switch cfg.output {
	case "print":
		printResults(results)
	case "raw":
		printRawResults(results)
	}
}
