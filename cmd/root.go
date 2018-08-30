package cmd

import (
	"github.com/felix-lessoer/qbeat/beater"

	cmd "github.com/elastic/beats/libbeat/cmd"
)

// Name of this beat
var Name = "qbeat"

// RootCmd to handle beats cli
var RootCmd = cmd.GenRootCmd(Name, "", beater.New)
