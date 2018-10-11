package beater

import (
	"strings"

	"github.com/elastic/beats/libbeat/logp"

	"github.com/ibm-messaging/mq-golang/ibmmq"
)

type Response struct {
	QueueName   string
	ChannelName string
	Metricset   string
	Metrictype  string
	MetricName  string
	Values      map[string]interface{}
}

var (
	err error
)

func connectLegacy(qMgrName string) error {
	qMgr, err := ibmmq.Conn(qMgrName)

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

	return err
}

func getQueueStatistics(localQueueName string) (map[string]*Response, error) {

	err = putCommand(ibmmq.MQCA_Q_NAME, localQueueName, ibmmq.MQCMD_RESET_Q_STATS)
	return parseResponse()
}

func getQueueStatus(localQueueName string) (map[string]*Response, error) {

	err = putCommand(ibmmq.MQCA_Q_NAME, localQueueName, ibmmq.MQCMD_INQUIRE_Q_STATUS)
	return parseResponse()
}

func getQueueMetadata(localQueueName string) (map[string]*Response, error) {

	err = putCommand(ibmmq.MQCA_Q_NAME, localQueueName, ibmmq.MQCMD_INQUIRE_Q)
	return parseResponse()
}

func getChannelStatus(channelName string) (map[string]*Response, error) {

	err = putCommand(ibmmq.MQCACH_CHANNEL_NAME, channelName, ibmmq.MQCMD_INQUIRE_CHANNEL_STATUS)
	return parseResponse()
}

/***
Translates a value that is delivered as number to a string
*/
func translateValue(key string, value int64) string {
	var mapping = make(map[string]string)
	mapping["mqiach_channel_status"] = "CHS"
	mapping["mqiach_channel_type"] = "CHT"
	mapping["mqia_q_type"] = "QT"

	if mapping[key] != "" {
		returnName := ibmmq.MQItoString(mapping[key], int(value))
		returnName = strings.ToLower(returnName)
		return returnName
	}
	return ""
}

func parseResponse() (map[string]*Response, error) {
	var buf = make([]byte, 131072)
	var elem *ibmmq.PCFParameter
	var responses map[string]*Response
	var resp *Response
	responses = make(map[string]*Response)
	// Loop here to get every message in the queue

	for err == nil {
		resp = new(Response)
		resp.Values = make(map[string]interface{})
		buf, err = GetMessageWithHObj(true, replyQObj)
		elemList, _ := ParsePCFResponse(buf)

		if err != nil {
			mqreturn := err.(*ibmmq.MQReturn)

			if mqreturn.MQCC == ibmmq.MQCC_FAILED && mqreturn.MQRC != ibmmq.MQRC_NO_MSG_AVAILABLE {
				logp.Debug("", "Error %v", err)
				return nil, err
			}
		}
		var key = ""
		for i := 0; i < len(elemList); i++ {
			elem = elemList[i]
			switch elem.Parameter {
			case ibmmq.MQCA_Q_NAME:
				for i := 0; i < len(elem.String); i++ {
					//logp.Debug("", "Current queue %v", strings.TrimSpace(elem.String[i]))
					resp.QueueName = strings.TrimSpace(elem.String[i])
					resp.Metricset = "queue"
					resp.Metrictype = "STATQ"
					key = resp.QueueName
				}
			case ibmmq.MQCACH_CHANNEL_NAME:
				for i := 0; i < len(elem.String); i++ {
					logp.Debug("", "Current channel %v", strings.TrimSpace(elem.String[i]))
					resp.ChannelName = strings.TrimSpace(elem.String[i])
					resp.Metricset = "channel"
					resp.Metrictype = "channelStatus"
					key = resp.ChannelName
				}
			default:
				if normalizeMetricNames(elem.Parameter) != "" {
					//logp.Debug("", "Current parameter (Type: %v): %v ",elem.Type, normalizeMetricNames(elem.Parameter))
					paramName := normalizeMetricNames(elem.Parameter)
					switch elem.Type {
					case ibmmq.MQCFT_INTEGER:
						resp.Values[paramName] = elem.Int64Value[0]
						logp.Debug("", "Try to translate %v: %v", paramName, elem.Int64Value[0])
						strValue := translateValue(paramName, elem.Int64Value[0])
						if strValue != "" {
							resp.Values[paramName+"_str"] = strValue
							logp.Debug("", "Translation successfull")
						}
					case ibmmq.MQCFT_STRING:
						resp.Values[paramName] = elem.String[0]
					default:
						//logp.Debug("", "Unhandeled parameter: %v", normalizeMetricNames(elem.Parameter))
					}
				}
			}
		}
		if key != "" {
			responses[key] = resp
		}
	}
	//Reset err if error is no more messages

	return responses, nil

}

func putCommand(paramType int32, paramValue string, commandCode int32) error {
	var buf []byte

	//Insert command
	putmqmd := ibmmq.NewMQMD()
	pmo := ibmmq.NewMQPMO()

	putmqmd.Format = "MQADMIN"
	putmqmd.ReplyToQ = replyQObj.Name
	putmqmd.MsgType = ibmmq.MQMT_REQUEST
	putmqmd.Report = ibmmq.MQRO_PASS_DISCARD_AND_EXPIRY

	// Reset QStats
	cfh := ibmmq.NewMQCFH()
	cfh.Command = commandCode

	logp.Info("%v for %v initiated", ibmmq.MQItoString("CMD", int(commandCode)), paramValue)

	// Add the parameters once at a time into a buffer
	pcfparm := new(ibmmq.PCFParameter)
	pcfparm.Type = ibmmq.MQCFT_STRING
	pcfparm.Parameter = paramType
	pcfparm.String = []string{paramValue}
	cfh.ParameterCount++
	buf = append(buf, pcfparm.Bytes()...)

	buf = append(cfh.Bytes(), buf...)

	// And put the command to the queue
	err = cmdQObj.Put(putmqmd, pmo, buf)

	if err != nil {
		logp.Info("Error putting the command into command queue")
		logp.Info("%v", err)
	}

	return err
}

func normalizeMetricNames(parameter int32) string {
	var returnName string
	returnName = ibmmq.MQItoString("IA", int(parameter))
	returnName = strings.ToLower(returnName)

	return returnName
}
