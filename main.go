package main

import (
	"os"

	"github.com/felix-lessoer/qbeat/cmd"

	_ "github.com/felix-lessoer/qbeat/beater"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
