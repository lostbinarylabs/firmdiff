package main

import (
	"os"

	"github.com/lostbinarylabs/firmdiff/internal/app"
)

func main() {
	if err := app.Execute(); err != nil {
		os.Exit(1)
	}
}
