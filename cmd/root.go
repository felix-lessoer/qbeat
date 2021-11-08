package cmd

import (
	"github.com/felix-lessoer/qbeat/beater"

	cmd "github.com/elastic/beats/v7/libbeat/cmd"

	"github.com/elastic/beats/v7/libbeat/cmd/instance"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/publisher/processing"
)

// Name of this beat
var Name = "qbeat"

// RootCmd to handle beats cli
var RootCmd *cmd.BeatsRootCmd

// withECSVersion is a modifier that adds ecs.version to events.
var withECSVersion = processing.WithFields(common.MapStr{
	"ecs": common.MapStr{
		"version": 1.11,
	},
})

// HeartbeatSettings contains the default settings for heartbeat
func QBeatSettings() instance.Settings {
	return instance.Settings{
		Name:          Name,
		Processing:    processing.MakeDefaultSupport(true, withECSVersion, processing.WithAgentMeta()),
		HasDashboards: false,
	}
}

// Initialize initializes the entrypoint commands for heartbeat
func Initialize(settings instance.Settings) *cmd.BeatsRootCmd {
	rootCmd := cmd.GenRootCmdWithSettings(beater.New, settings)

	return rootCmd
}

func init() {
	RootCmd = Initialize(QBeatSettings())
}
