package main

import (
	"os"

	"github.com/elastic/beats/libbeat/cmd"
	"github.com/elastic/beats/libbeat/cmd/instance"

	"github.com/felix-lessoer/qbeat/beater"
)

var RootCmd = cmd.GenRootCmdWithSettings(beater.New, instance.Settings{Name: "qbeat"})

func main() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
