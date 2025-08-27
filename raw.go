package main

import (
	"fmt"
)

func printRawResults(results *results) {
	if cfg.status {
		fmt.Println(results.status.v.Raw)
	}
	if cfg.bedrock.enabled {
		fmt.Println(results.bedrock.v.Raw)
	}
	if cfg.query.enabled {
		fmt.Println(results.query.v.Raw)
	}
}
