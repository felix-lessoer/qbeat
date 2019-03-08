// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import (
	"time"
)

type ConnectionConfig struct {
	ClientMode bool   `config:"clientMode"`
	MqServer   string `config:"mqServer"`
	UserId     string `config:"userId"`
	Password   string `config:"password"`
}

type Config struct {
	Period             time.Duration    `config:"period"`
	QueueManager       string           `config:"bindingQueueManager"`
	RemoteQueueManager []string         `config:"targetQueueManager"`
	LocalQueue         string           `config:"queue"`
	QueueStatus        bool             `config:"queueStatus"`
	QueueStats         bool             `config:"queueStats"`
	Channel            string           `config:"channel"`
	QMgrStat           bool             `config:"queueManagerStatus"`
	PubSub             bool             `config:"pubSub"`
	Advanced           string           `config:"advanced"`
	CC                 ConnectionConfig `config:"cc"`
}

var (
	DefaultConfig = Config{
		PubSub:             false,
		QMgrStat:           true,
		RemoteQueueManager: []string{""},
		LocalQueue:         "*",
		QueueStatus:        true,
		QueueStats:         true,
		Channel:            "*",
		Advanced:           "",
		CC: ConnectionConfig{
			ClientMode: false,
			UserId:     "",
			Password:   "",
		},
	}
)
