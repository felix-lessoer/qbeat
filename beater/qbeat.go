package beater

import (
	"fmt"
	"time"
	"strings"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/felix-lessoer/qbeat/config"

	"github.com/ibm-messaging/mq-golang/ibmmq"
	"github.com/ibm-messaging/mq-golang/mqmetric"
)

// Qbeat configuration.
type Qbeat struct {
	done   chan struct{}
	config config.Config
	client beat.Client
}

var (
	first      = true
	errorCount = 0
	cmdQObj   ibmmq.MQObject
	replyQObj ibmmq.MQObject
	qMgr ibmmq.MQQueueManager
	targetQueues []string
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
	err = mqmetric.InitConnection(bt.config.QueueManager, "SYSTEM.DEFAULT.MODEL.QUEUE", &bt.config.CC)
	if err == nil {
		logp.Info("Connected to queue manager %v", bt.config.QueueManager)
		defer mqmetric.EndConnection()
	}

	logp.Info("Connect to QM done")

	// What metrics can the queue manager provide? Find out, and
	// subscribe.
	if err == nil {
		logp.Info("DiscoverAndSubscribe start")
		err = mqmetric.DiscoverAndSubscribe(bt.config.LocalQueue, true, "")
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

	for _, cl := range mqmetric.Metrics.Classes {
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
	mqmetric.ProcessPublications()
	logp.Info("ProcessPublications done")

	if first {
		// Always ignore the first loop through as there might
		// be accumulated stuff from a while ago, and lead to
		// a misleading range on graphs.
		first = false
		logp.Info("First loop done")
	} else {
	logp.Info("Start second loop")
	firstPoint := true

	for _, cl := range mqmetric.Metrics.Classes {
		for _, ty := range cl.Types {
			event := beat.Event{
				Timestamp: time.Now(),
				Fields: common.MapStr{
					"metrictype": cl.Name,
					"objecttopic": ty.ObjectTopic,
					"type":    b.Info.Name,
					"qmgr": bt.config.QueueManager,
				},
			}
			for _, elem := range ty.Elements {
				for key, value := range elem.Values {
					if firstPoint {
						firstPoint = false
					}
					f := mqmetric.Normalise(elem, key, value)

					//Add some metadata information based on type
					if key != mqmetric.QMgrMapKey {
							event.Fields["queue"] = key;
							event.Fields["metricset"] = "queue";
					} else {
							event.Fields["metricset"] = "queuemanager";
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

func connectLegacy(bt *Qbeat)(error) {
	logp.Info("Connect in legacy mode")
	qMgr, err := ibmmq.Conn(bt.config.QueueManager)

	logp.Info("Connect to command queue")
	//Connect to Command Queue
	mqod := ibmmq.NewMQOD()
	openOptions := ibmmq.MQOO_OUTPUT | ibmmq.MQOO_FAIL_IF_QUIESCING
	mqod.ObjectType = ibmmq.MQOT_Q
	mqod.ObjectName = "SYSTEM.ADMIN.COMMAND.QUEUE"
	cmdQObj, err = qMgr.Open(mqod, openOptions)

	if err != nil {
		return err
	}

	logp.Info("Connect to Reply queue")
	//Connect to Reply Queue
	mqod2 := ibmmq.NewMQOD()
	openOptions2 := ibmmq.MQOO_INPUT_AS_Q_DEF | ibmmq.MQOO_FAIL_IF_QUIESCING
	mqod2.ObjectType = ibmmq.MQOT_Q
	mqod2.ObjectName = "SYSTEM.DEFAULT.MODEL.QUEUE"
	replyQObj, err = qMgr.Open(mqod2, openOptions2)

	if err != nil {
		return err
	}

	//#########################
	//###Load queue names
	//#########################

	var buf []byte
	var elem *ibmmq.PCFParameter
	var datalen int

	// Can allow all the other fields to default
	putmqmd := ibmmq.NewMQMD()
	pmo := ibmmq.NewMQPMO()

	pmo.Options = ibmmq.MQPMO_NO_SYNCPOINT
	pmo.Options |= ibmmq.MQPMO_NEW_MSG_ID
	pmo.Options |= ibmmq.MQPMO_NEW_CORREL_ID
	pmo.Options |= ibmmq.MQPMO_FAIL_IF_QUIESCING

	putmqmd.Format = "MQADMIN"
	putmqmd.ReplyToQ = replyQObj.Name
	putmqmd.MsgType = ibmmq.MQMT_REQUEST
	putmqmd.Report = ibmmq.MQRO_PASS_DISCARD_AND_EXPIRY

	cfh := ibmmq.NewMQCFH()
	cfh.Command = ibmmq.MQCMD_INQUIRE_Q_NAMES

	// Add the parameters one at a time into a buffer
	pcfparm := new(ibmmq.PCFParameter)
	pcfparm.Type = ibmmq.MQCFT_STRING
	pcfparm.Parameter = ibmmq.MQCA_Q_NAME
	pcfparm.String = []string{bt.config.LocalQueue}
	cfh.ParameterCount++
	buf = append(buf, pcfparm.Bytes()...)

	pcfparm = new(ibmmq.PCFParameter)
	pcfparm.Type = ibmmq.MQCFT_INTEGER
	pcfparm.Parameter = ibmmq.MQIA_Q_TYPE
	pcfparm.Int64Value = []int64{int64(ibmmq.MQQT_LOCAL)}
	cfh.ParameterCount++
	buf = append(buf, pcfparm.Bytes()...)

	// Once we know the total number of parameters, put the
	// CFH header on the front of the buffer.
	buf = append(cfh.Bytes(), buf...)

	// And put the command to the queue
	err = cmdQObj.Put(putmqmd, pmo, buf)

	if err != nil {
		return err
	}

	// Now get the response
	getmqmd := ibmmq.NewMQMD()
	gmo := ibmmq.NewMQGMO()
	gmo.Options = ibmmq.MQGMO_NO_SYNCPOINT
	gmo.Options |= ibmmq.MQGMO_FAIL_IF_QUIESCING
	gmo.Options |= ibmmq.MQGMO_WAIT
	gmo.Options |= ibmmq.MQGMO_CONVERT
	gmo.WaitInterval = 30 * 1000

	// Ought to add a loop here in case we get truncated data
	buf = make([]byte, 32768)

	datalen, err = replyQObj.Get(getmqmd, gmo, buf)
	if err != nil {
		 return err
	 }


		cfh, offset := ibmmq.ReadPCFHeader(buf)
		if cfh.CompCode != ibmmq.MQCC_OK {
			fmt.Errorf("PCF command failed with CC %d RC %d", cfh.CompCode, cfh.Reason)
		} else {
			parmAvail := true
			bytesRead := 0
			for parmAvail && cfh.CompCode != ibmmq.MQCC_FAILED {
				elem, bytesRead = ibmmq.ReadPCFParameter(buf[offset:])
				offset += bytesRead
				// Have we now reached the end of the message
				if offset >= datalen {
					parmAvail = false
				}

				switch elem.Parameter {
				case ibmmq.MQCACF_Q_NAMES:
					if len(elem.String) == 0 {
						fmt.Errorf("No queues matching '%s' exist", bt.config.LocalQueue)
					}
					for i := 0; i < len(elem.String); i++ {
						logp.Info("add queue %v",strings.TrimSpace(elem.String[i]) )
						targetQueues = append(targetQueues, strings.TrimSpace(elem.String[i]))
					}
				}
			}
		}

	logp.Info("Connection successfull")
	return err
}

func collectLegacy(bt *Qbeat, b *beat.Beat)(error) {
	//Collect queue statistics
	var buf []byte
	var err error
	var elem *ibmmq.PCFParameter
	var datalen int

		logp.Info("Collect in legacy mode")
		logp.Info("Queues %v", targetQueues)
for i := 0; i < len(targetQueues); i++ {
		//Reset variables for the next loop
		err = nil
		buf = nil

		event := beat.Event{
			Timestamp: time.Now(),
			Fields: common.MapStr{
				"type": b.Info.Name,
				"qmgr": bt.config.QueueManager,
			},
		}


		logp.Info("##################")
		logp.Info("Insert new command")
		//Insert command
		putmqmd := ibmmq.NewMQMD()
		pmo := ibmmq.NewMQPMO()

		//pmo.Options = ibmmq.MQCMD_RESET_Q_STATS

		putmqmd.Format = "MQADMIN"
		putmqmd.ReplyToQ = replyQObj.Name
		putmqmd.MsgType = ibmmq.MQMT_REQUEST
		putmqmd.Report = ibmmq.MQRO_PASS_DISCARD_AND_EXPIRY

		// Reset QStats
		cfh := ibmmq.NewMQCFH()
		cfh.Command = ibmmq.MQCMD_RESET_Q_STATS

		logp.Info("MQCMD_RESET_Q_STATS for %v initiated", targetQueues[i])

		// Add the parameters one at a time into a buffer
		pcfparm := new(ibmmq.PCFParameter)
		pcfparm.Type = ibmmq.MQCFT_STRING
		pcfparm.Parameter = ibmmq.MQCA_Q_NAME
		//pcfparm.String = []string{bt.config.LocalQueue}
		pcfparm.String = []string{"*"}
		//pcfparm.String = targetQueues
		cfh.ParameterCount++
		buf = append(buf, pcfparm.Bytes()...)

		buf = append(cfh.Bytes(), buf...)

		// And put the command to the queue
		err = cmdQObj.Put(putmqmd, pmo,  buf)

		if err != nil {
			logp.Info("Error putting the command into command queue")
			logp.Info("%v", err)
			continue
		}

		// Now get the response
		getmqmd := ibmmq.NewMQMD()
		gmo := ibmmq.NewMQGMO()
		//gmo.Options = ibmmq.MQGMO_NO_SYNCPOINT
		gmo.Options = ibmmq.MQGMO_FAIL_IF_QUIESCING
		gmo.Options |= ibmmq.MQGMO_WAIT
		gmo.Options |= ibmmq.MQGMO_CONVERT
		gmo.WaitInterval = 10 * 1000

		// Ought to add a loop here in case we get truncated data
			buf = make([]byte, 32768)
			datalen, err = replyQObj.Get(getmqmd, gmo, buf)

			// Always ignore the first loop through as there might
			// be accumulated stuff from a while ago, and lead to
			// a misleading range on graphs.
			if !first {
			if err == nil {
				cfh, offset := ibmmq.ReadPCFHeader(buf)
				if cfh.CompCode != ibmmq.MQCC_OK {
					logp.Info("PCF command failed")
				} else {
					logp.Info("PCF command successfull")
					parmAvail := true
					bytesRead := 0
					for parmAvail && cfh.CompCode != ibmmq.MQCC_FAILED {
						elem, bytesRead = ibmmq.ReadPCFParameter(buf[offset:])
						offset += bytesRead
						// Have we now reached the end of the message
						if offset >= datalen {
							parmAvail = false
						}

						if elem.Parameter == ibmmq.MQCA_Q_NAME {
							if len(elem.String) == 0 {
								logp.Info("No queues matching %v", targetQueues[i])
							}
							for i := 0; i < len(elem.String); i++ {
								event.Fields["queue"] = strings.TrimSpace(elem.String[i]);
								event.Fields["metricset"] = "queue";
								event.Fields["metrictype"] = "STATQ";
							}
						} else {
								if (elem.Type == ibmmq.MQCFT_INTEGER) {
									event.Fields[ibmmq.MQItoString("IA", int(elem.Parameter))] = elem.Int64Value[0]
								} else {
									event.Fields[ibmmq.MQItoString("IA", int(elem.Parameter))] = elem.String
								}
						}
					}
				}
			} else {
				logp.Info("There was an error")
				return err
			}

		bt.client.Publish(event)
}
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

	//Set the mode based on config
	var legacy bool
	if bt.config.Mode == "PubSub" {
			legacy = false
	}
	if bt.config.Mode == "Legacy" {
			legacy = true
	}


	if (legacy) {
		err = connectLegacy(bt)
	} else {
		err = connectPubSub(bt)
	}

	ticker := time.NewTicker(bt.config.Period)
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

		if (legacy) {
			collectLegacy(bt, b)
		} else {
			collectPubSub(bt, b)
		}

		//This is to ignore the first chunk of data because this can have inappriate data
		first = false;

		logp.Info("Events sent")
	}
}

// Stop stops qbeat.
func (bt *Qbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
