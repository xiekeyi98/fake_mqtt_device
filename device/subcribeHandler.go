package device

import (
	"encoding/json"
	"fmt"
	"strings"
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
	switch received.Method {
	case "report_reply":
		logrus.Debugf("report_reply,ignore.")
		return
	case "control":
		publish(client, strings.Replace(message.Topic(), "down", "up", 1), fmt.Sprintf(`{"method":"control_reply","clientToken":"%s","code":0,"status":"succ"}`, sent.ClientToken))
		logger.Infof("control_reply")
	}
	// report
	sent.ClientToken = received.ClientToken
	sent.Method = "report"
	sent.Params = received.Params
	sent.Timestamp = time.Now().Unix()
	stBytes, _ := json.Marshal(sent)
	publish(client, strings.Replace(message.Topic(), "down", "up", 1), string(stBytes))
	logger.Infof("report")
}
