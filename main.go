package main

import (
	"os"

	"github.com/felix-lessoer/qbeat/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
