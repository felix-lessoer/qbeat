// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import (
	"time"
)

type ConnectionConfig struct {
	ClientMode bool
	UserId     string
	Password   string
}

type Config struct {
	Period       time.Duration `config:"period"`
	QueueManager string        `config:"queueManager"`
	LocalQueue   string        `config:"localQueue"`
	Channel      string        `config:"channel"`
	PubSub       bool          `config:"pubSub"`
	CC           ConnectionConfig
}

var (
	DefaultConfig = Config{
		PubSub:     false,
		LocalQueue: "*",
		Channel:    "*",
	}
)
