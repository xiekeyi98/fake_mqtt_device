package device

import (
	"encoding/json"
	"fmt"
	"strings"
	"testUtils/fakeDevice/clog"
	"testUtils/fakeDevice/config"
	"testUtils/fakeDevice/device/ota"
	"testUtils/fakeDevice/utils"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// 订阅消息的抓手!
func (resp *DeviceCtx) subHandler(client mqtt.Client, message mqtt.Message) {
	payloadStr := string(message.Payload())
	newCtx := clog.WithCtx(resp.ctx, "topic", message.Topic())
	if resp.IsSystemTopic(message.Topic()) {
		clog.Logger(newCtx).Infof("收到消息:\n%v\n", utils.GetPrettyJSONStr(payloadStr))
	} else {
		clog.Logger(newCtx).Infof("收到非系统 topic 消息:[%s]", payloadStr)
		return
	}

	if resp.Device.IsShadow() {
		resp.HandlerIoTHubProtocol(message)
	} else {
		resp.HandlerExplorerProtocol(message)
	}

}

func (resp *DeviceCtx) HandlerIoTHubProtocol(message mqtt.Message) {
	var (
		received = shadowPayload{}
	)
	if err := json.Unmarshal(message.Payload(), &received); err != nil {
		clog.Logger(resp.ctx).WithError(err).Errorf("JSON 解码失败")
	}
	switch received.Type {
	case "update_firmware":
		ota.NewOTATask(resp.ctx, resp.ProductId, resp.DeviceName, resp).OTAFirmwareUpdate(string(message.Payload()))
	case "get":
		resp.OnShadowDown(received)
	case "delta":
		resp.OnShadowDown(received)
	default:
		clog.Logger(resp.ctx).Warnf("尚不支持的方法:%s", received.Type)
	}
}

func (resp *DeviceCtx) HandlerExplorerProtocol(message mqtt.Message) {
	var (
		received = ExplorerPayload{}
		sent     = ExplorerPayload{}
	)
	if err := json.Unmarshal(message.Payload(), &received); err != nil {
		clog.Logger(resp.ctx).WithError(err).Errorf("JSON 解码失败")
	}
	switch received.GetMethodOrType() {
	case "update_firmware":
		ota.NewOTATask(resp.ctx, resp.ProductId, resp.DeviceName, resp).OTAFirmwareUpdate(string(message.Payload()))
		//resp.OTAReport(string(message.Payload()))
	case "report_reply", "event_reply",
		"report_version_rsp":
		return
	case "get_status_reply":
		clog.Logger(resp.ctx).Debugf("收到 get_status_reply 消息,主动上报一次当前状态。")
		resp.sendReportProperty(received.ClientToken, changeTopicUP2Down(message.Topic()), resp.getStatusReplyReportedParams(string(message.Payload())))

	case "control":
		resp.Publish(
			strings.Replace(message.Topic(), "down", "up", 1),
			[]byte(fmt.Sprintf(`{"method":"control_reply","clientToken":"%s","code":0,"status":"succ"}`,
				sent.ClientToken)))
		clog.Logger(resp.ctx).Infof("发送 control_reply 消息")
		resp.sendReportProperty(received.ClientToken, changeTopicUP2Down(message.Topic()), received.Params)
	case "action":
		resp.sendActionReply(received.ClientToken, changeTopicUP2Down(message.Topic()), received.ActionId)
	default:
		clog.Logger(resp.ctx).Warnf("尚不支持的方法:%s", received.GetMethodOrType())
	}
}

func changeTopicUP2Down(topic string) string {
	return strings.Replace(topic, "down", "up", 1)
}

func (resp *DeviceCtx) sendReportProperty(clientToken, topic string, Params map[string]interface{}) {
	sent := ExplorerPayload{}
	sent.ClientToken = clientToken
	sent.Method = "report"
	sent.Params = Params
	sent.Timestamp = time.Now().Unix()
	stBytes, _ := json.Marshal(sent)
	resp.Publish(topic, stBytes)
	clog.Logger(resp.ctx).Infof("上报属性信息.")
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
		clog.Logger(resp.ctx).WithError(err).Warnf("获取 ActionId 失败")
	}
	stBytes, _ := json.Marshal(sent)
	resp.Publish(topic, stBytes)
	clog.Logger(resp.ctx).Infof("上报 action_reply 消息。")
}

func (resp *DeviceCtx) getStatusReplyReportedParams(getStatusReply string) map[string]interface{} {
	type statusReply struct {
		Method      string `json:"method"`
		ClientToken string `json:"clientToken"`
		Code        int    `json:"code"`
		Status      string `json:"status"`
		Type        string `json:"report"`
		Data        struct {
			Report map[string]interface{} `json:"reported"`
		} `json:"data"`
	}
	reply := statusReply{}
	if err := json.Unmarshal([]byte(getStatusReply), &reply); err != nil {
		clog.Logger(resp.ctx).WithError(err).Warnf("JSON 反序列化失败")
	}
	return reply.Data.Report

}

func (resp *DeviceCtx) IsSystemTopic(topic string) bool {
	return strings.HasPrefix(topic, "$")

}
