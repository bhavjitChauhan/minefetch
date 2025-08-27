package main

import (
	"log"
)

func main() {
	log.SetFlags(0)

	err := parseArgs()
	if err != nil {
		log.Fatalln("Failed to parse arguments:", err)
	}

	results := getResults()

	switch cfg.output {
	case "print":
		printResults(results)
	case "raw":
		printRawResults(results)
	}
}
