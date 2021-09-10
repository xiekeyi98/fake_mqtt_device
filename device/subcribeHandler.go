package device

import (
	"encoding/json"
	"fmt"
	"strings"
	"testUtils/fakeDevice/config"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
)

// 订阅消息的抓手!
func (resp *DeviceCtx) subHandler(client mqtt.Client, message mqtt.Message) {
	payloadStr := string(message.Payload())
	logger := logrus.StandardLogger()
	logger.Infof("receive: %v", payloadStr)
	var (
		received = Payload{}
		sent     = Payload{}
	)

	if err := json.Unmarshal(message.Payload(), &received); err != nil {
		logger.Errorf("err:%v", err)
	}
	logger = logger.WithFields(
		map[string]interface{}{
			"ProductId":  resp.ProductId,
			"DeviceName": resp.DeviceName,
		},
	).Logger

	switch received.GetMethodOrType() {
	case "update_firmware":
		resp.OTAReport(string(message.Payload()))
	case "report_reply", "event_reply",
		"get_status_reply", "report_version_rsp":
		logrus.Debugf("%s,ignore.", received.GetMethodOrType())
		return
	case "control":
		publish(client,
			strings.Replace(message.Topic(), "down", "up", 1),
			fmt.Sprintf(`{"method":"control_reply","clientToken":"%s","code":0,"status":"succ"}`,
				sent.ClientToken))
		logger.Infof("control_reply")
		resp.sendReportProperty(received.ClientToken, changeTopicUP2Down(message.Topic()), received.Params)
	case "action":
		resp.sendActionReply(received.ClientToken, changeTopicUP2Down(message.Topic()), received.ActionId)
	default:
		logger.Warnf("get unsupported method:[%s]", received.Method)
	}
}
func changeTopicUP2Down(topic string) string {
	return strings.Replace(topic, "down", "up", 1)
}

func (resp *DeviceCtx) sendReportProperty(clientToken, topic string, Params map[string]interface{}) {
	sent := Payload{}
	sent.ClientToken = clientToken
	sent.Method = "report"
	sent.Params = Params
	sent.Timestamp = time.Now().Unix()
	stBytes, _ := json.Marshal(sent)
	publish(resp.MQTTClient, topic, string(stBytes))
	logrus.Infof("report property")
}

func (resp *DeviceCtx) sendActionReply(clientToken, topic string, actionId string) {
	type ActionReply struct {
		Method      string                 `json:"method"`
		ClientToken string                 `json:"clientToken"`
		Code        int                    `json:"code"`
		Status      string                 `json:"status"`
		Response    map[string]interface{} `json:"response"`
	}

	var (
		sent = ActionReply{}
		err  error
	)
	sent.ClientToken = clientToken
	sent.Method = "action_reply"
	sent.Code = 0
	sent.Status = "succ"
	sent.Response, err = config.GetActionParams(actionId)
	if err != nil {
		logrus.Warnf("get action err:%v", err)
	}
	stBytes, _ := json.Marshal(sent)
	publish(resp.MQTTClient, topic, string(stBytes))
	logrus.Infof("report property")
}
