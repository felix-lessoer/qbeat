package beater

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/felix-lessoer/qbeat/config"
)

// Qbeat configuration.
type Qbeat struct {
	done   chan struct{}
	config config.Config
	client beat.Client
}

type RequestObject struct {
	Commands []struct {
		Cmd    string
		Params map[string]interface{}
	}
}

var (
	first      = true
	errorCount = 0
)

// New creates an instance of qbeat.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	c := config.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Qbeat{
		done:   make(chan struct{}),
		config: c,
	}
	return bt, nil
}

func connectPubSub(bt *Qbeat) error {
	var err error

	bt.config.CC.ClientMode = false

	// Connect to MQ

	logp.Info("Connect to QM %v start", bt.config.QueueManager)
	err = InitConnection(bt.config.QueueManager, "SYSTEM.DEFAULT.MODEL.QUEUE", &bt.config.CC)
	if err == nil {
		logp.Info("Connected to queue manager %v", bt.config.QueueManager)
	}

	logp.Info("Connect to QM done")

	// What metrics can the queue manager provide? Find out, and
	// subscribe.
	if err == nil {
		logp.Info("DiscoverAndSubscribe start")
		err = DiscoverAndSubscribe(bt.config.LocalQueue, true, "")
	}
	logp.Info("DiscoverAndSubscribe done")

	return err
}

func collectPubSub(bt *Qbeat, b *beat.Beat) {
	// #####Code for collecting the MQ metrics
	// Clear out everything we know so far. In particular, replace
	// the map of values for each object so the collection starts
	// clean.
	logp.Info("Start MQ Metric collection")

	for _, cl := range Metrics.Classes {
		//logp.Info("Define class %v", cl.Name)
		for _, ty := range cl.Types {
			//logp.Info("Define type %v", ty.ObjectTopic)
			for _, elem := range ty.Elements {
				//logp.Info("Define elem %v", elem.Values)
				//logp.Info("test: ",elem.Values)
				elem.Values = make(map[string]int64)
			}
		}
	}

	//if (cl.length > 0) {
	// Process all the publications that have arrived
	logp.Info("ProcessPublications start")
	ProcessPublications()
	logp.Info("ProcessPublications done")

	if !first {

		for _, cl := range Metrics.Classes {
			for _, ty := range cl.Types {
				event := beat.Event{
					Timestamp: time.Now(),
					Fields: common.MapStr{
						"metrictype":  cl.Name,
						"objecttopic": ty.ObjectTopic,
						"type":        b.Info.Name,
						"qmgr":        bt.config.QueueManager,
					},
				}
				for _, elem := range ty.Elements {
					for key, value := range elem.Values {
						f := Normalise(elem, key, value)

						//Add some metadata information based on type
						if key != QMgrMapKey {
							event.Fields["queue"] = key
							event.Fields["metricset"] = "queue"
						} else {
							event.Fields["metricset"] = "queuemanager"
						}
						event.Fields[elem.MetricName] = float32(f)
					}
				}
				bt.client.Publish(event)
			}
		}

		//}
	}
	// ###### END Code for collecting the MQ metrics

}

func connectLegacyMode(bt *Qbeat) error {
	logp.Info("Connect in legacy mode")

	err = connectLegacy(bt.config.QueueManager, bt.config.RemoteQueueManager)

	if err != nil {
		return err
	}

	logp.Info("Connection successfull")
	return err
}

func createEvents(eventType string, qmgrName string, responseObj map[string]*Response) []beat.Event {
	var events []beat.Event

	for id, elem := range responseObj {
		event := beat.Event{
			Timestamp: time.Now(),
			Fields: common.MapStr{
				"type":         eventType,
				"qmgr":         qmgrName,
				"metricset":    elem.Metricset,
				"metrictype":   elem.Metrictype,
				"targetObject": elem.TargetObject,
			},
		}
		for key, value := range responseObj[id].Values {
			event.Fields[key] = value
		}

		events = append(events, event)
	}
	return events
}

func mergeEventsWithResponseObj(events []beat.Event, responseObj map[string]*Response) []beat.Event {

	for id, _ := range responseObj {
		for _, event := range events {
			if id == event.Fields["targetObject"] {
				for key, value := range responseObj[id].Values {
					event.Fields[key] = value
				}
			}
		}
	}
	return events
}

func generateConnectedObjectsField(events []beat.Event) []beat.Event {
	for _, event := range events {
		var connections []string
		connections = make([]string, 0)
		connections = append(connections, event.Fields["qmgr"].(string))
		connections = append(connections, event.Fields["targetObject"].(string))
		if event.Fields["mqcach_xmit_q_name"] != nil {
			connections = append(connections, event.Fields["mqcach_xmit_q_name"].(string))
		}
		if event.Fields["mqca_remote_q_mgr_name"] != nil {
			connections = append(connections, event.Fields["mqca_remote_q_mgr_name"].(string))
		}
		if event.Fields["mqca_remote_q_name"] != nil {
			connections = append(connections, event.Fields["mqca_remote_q_name"].(string))
		}
		//Remove whitespaces
		for _, connection := range connections {
			connection = strings.TrimSpace(connection)
		}
		event.Fields["Conntections"] = connections
	}
	return events
}

func collectLegacy(bt *Qbeat, b *beat.Beat) error {
	//Collect queue statistics
	var err error
	var events []beat.Event

	if bt.config.Advanced != "" {
		logp.Info("Start collecting in advance object")
		var requestObject RequestObject
		err := json.Unmarshal([]byte(bt.config.Advanced), &requestObject)

		if err != nil {
			return err
		}
		logp.Info("Advanced json: %v", requestObject)
		for _, command := range requestObject.Commands {
			responseObj, err := getAdvancedResponse(command.Cmd, command.Params)

			if err != nil {
				return err
			}

			events = append(events, createEvents(b.Info.Name, bt.config.QueueManager, responseObj)...)
		}
	}
	if bt.config.QMgrStat {
		qMgrMetadata, err := getQManagerMetadata()
		if err != nil {
			return err
		}

		qMgrStatus, err := getQManagerStatus()
		if err != nil {
			return err
		}
		tmpEvents := createEvents(b.Info.Name, bt.config.QueueManager, qMgrMetadata)
		events = append(events, mergeEventsWithResponseObj(tmpEvents, qMgrStatus)...)
	}

	if bt.config.Channel != "" {
		chMetadata, err := getChannelMetadata(bt.config.Channel)
		if err != nil {
			return err
		}
		chStatus, err := getChannelStatus(bt.config.Channel)
		if err != nil {
			return err
		}

		tmpEvents := createEvents(b.Info.Name, bt.config.QueueManager, chMetadata)
		events = append(events, mergeEventsWithResponseObj(tmpEvents, chStatus)...)
	}

	if bt.config.LocalQueue != "" {
		qMetadata, err := getQueueMetadata(bt.config.LocalQueue)
		if err != nil {
			return err
		}

		qStatus, err := getQueueStatus(bt.config.LocalQueue)
		if err != nil {
			return err
		}

		qStatistics, err := getQueueStatistics(bt.config.LocalQueue)
		if err != nil {
			return err
		}
		tmpEvents := createEvents(b.Info.Name, bt.config.QueueManager, qMetadata)
		tmpEvents = mergeEventsWithResponseObj(tmpEvents, qStatus)
		events = append(events, mergeEventsWithResponseObj(tmpEvents, qStatistics)...)
	}

	//Add a field that contains all connections to other MQ objects
	// this helps to visualize via graph
	events = generateConnectedObjectsField(events)

	// Always ignore the first loop through as there might
	// be accumulated stuff from a while ago, and lead to
	// a misleading range on graphs.
	if !first {
		bt.client.PublishAll(events)
	}

	return err
}

// Run starts qbeat.
func (bt *Qbeat) Run(b *beat.Beat) error {
	logp.Info("qbeat is running! Hit CTRL-C to stop it.")

	var err error

	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	var legacy = false
	if bt.config.LocalQueue != "" || bt.config.Channel != "" || bt.config.QMgrStat || bt.config.Advanced != "" {
		legacy = true
	}

	if legacy {
		err = connectLegacyMode(bt)
		if err != nil {
			logp.Info("Wasn't able to connect due to an error")
			return err
		}
	}

	if bt.config.PubSub {
		err = connectPubSub(bt)

		if err != nil {
			logp.Info("Wasn't able to connect due to an error")
			return err
		}
	}

	ticker := time.NewTicker(bt.config.Period)
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

		if legacy {
			err = collectLegacy(bt, b)
		}
		if bt.config.PubSub {
			collectPubSub(bt, b)
		}

		if err != nil {
			return err
		}

		//This is to ignore the first chunk of data because this can have inappropiate data
		if first {
			first = false
			logp.Info("Events ignored in the first loop")
			continue
		}

		logp.Info("Events sent")
	}
}

// Stop stops qbeat.
func (bt *Qbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
