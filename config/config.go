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
	Period             time.Duration `config:"period"`
	QueueManager       string        `config:"queueManager"`
	RemoteQueueManager string        `config:"remoteQueueManager"`
	LocalQueue         string        `config:"localQueue"`
	Channel            string        `config:"channel"`
	QMgrStat           bool          `config:"queueManagerStatus"`
	PubSub             bool          `config:"pubSub"`
	Advanced           string        `config:"advanced"`
	CC                 ConnectionConfig
}

var (
	DefaultConfig = Config{
		PubSub:             false,
		QMgrStat:           true,
		RemoteQueueManager: "",
		LocalQueue:         "*",
		Channel:            "*",
		Advanced:           "",
	}
)
