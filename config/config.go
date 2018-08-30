// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import (
	"time"

	"github.com/ibm-messaging/mq-golang/mqmetric"
)

type Config struct {
	Period time.Duration `config:"period"`
	QueueManager  string       `config:"queueManager"`
	LocalQueue 		string       `config:"localQueue"`
	Mode 					string       `config:"mode"`
	CC						mqmetric.ConnectionConfig
}

var (
	DefaultConfig = Config{
		Mode:	"PubSub",
		LocalQueue: "*",
	}
)
