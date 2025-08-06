package main

import (
	"fmt"
	"minefetch/internal/mc"
	"minefetch/internal/mcpe"
)

func printRawResults(results results) {
	print := func(result result) {
		switch v := result.v.(type) {
		case mc.StatusResponse:
			fmt.Println(v.Raw)
		case mcpe.StatusResponse:
			fmt.Println(v.Raw)
		case mc.QueryResponse:
			fmt.Println(v.Raw)
		default:
			fmt.Println()
		}
	}

	if cfg.status {
		print(results[resultStatus])
	}
	if cfg.bedrock || cfg.crossplay {
		print(results[resultBedrockStatus])
	}
	if cfg.query {
		print(results[resultQuery])
	}
}
