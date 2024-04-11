package main

import (
	"log"
	"os"

	"github.com/grafana/nethax/cmd"
)

func main() {
	err := cmd.Execute(os.Args[1:])
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
}
