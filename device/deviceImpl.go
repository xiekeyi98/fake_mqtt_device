package device

import (
	"context"
	"fmt"
	"testUtils/fakeDevice/clog"
	"testUtils/fakeDevice/config"
	"testUtils/fakeDevice/device/shadow"
	"testUtils/fakeDevice/utils"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

// 设备连接的上下文
type DeviceCtx struct {
	mqttClient mqtt.Client // MQTT 客户端

	config.Device // 设备的配置

	ctx       context.Context
	cancelCtx context.CancelFunc

	shadow shadow.ShadowInfo
}

func GetDeviceCtx(ctx context.Context, d config.Device) (DeviceInterface, error) {
	ctx, cancel := context.WithCancel(ctx)
	ctx = clog.WithCtx(ctx, "ProductId", d.ProductId)
	ctx = clog.WithCtx(ctx, "DeviceName", d.DeviceName)
	resp := DeviceCtx{
		ctx:       ctx,
		cancelCtx: cancel,
		Device:    d,
	}
	mqttDSN, err := d.GetMQTTDSN()
	if err != nil {
		return nil, err
	}

	mqttClientOpts := mqtt.
		NewClientOptions().
		SetOrderMatters(false).
		AddBroker(fmt.Sprintf("tcp://%s:%d", mqttDSN.Broker, mqttDSN.Port)).
		SetClientID(fmt.Sprintf("%s%s", d.ProductId, d.DeviceName)).
		SetUsername(mqttDSN.UserName).
		SetPassword(mqttDSN.Token)

	mqttClientOpts.OnConnectionLost = func(client mqtt.Client, err error) {
		clog.Logger(resp.ctx).WithError(err).Errorf("connect lost.")
	}
	client := mqtt.NewClient(mqttClientOpts)
	resp.mqttClient = client
	return &resp, nil
}

func (resp *DeviceCtx) Connect() error {
	token := resp.mqttClient.Connect()
	if wa, err := token.Wait(), token.Error(); !wa || err != nil {
		clog.Logger(resp.ctx).WithError(err).Errorf("等待 MQTT 完成错误")
		return errors.Cause(err)
	}
	if err := resp.subAllTopics(); err != nil {
		clog.Logger(resp.ctx).WithError(err).Errorf("订阅 topic 失败")
		return errors.Cause(err)
	}
	clog.Logger(resp.ctx).Infof("连接并订阅 topic 成功。")
	resp.postConnect()
	return nil
}

func (resp *DeviceCtx) postConnect() {
	if resp.Device.IsShadow() {
		clog.Logger(resp.ctx).Debugf("使用 hub 影子协议")
		resp.ShadowGet()
	} else {
		resp.GetStatus()
		resp.ReportOTAVersion(resp.Device.DeviceVersion)
		go resp.ReportEvents()

	}

}

func (resp *DeviceCtx) Disconnect() {
	resp.cancelCtx()
	resp.mqttClient.Disconnect(2000)
}

type ExplorerPayload struct {
	Method      string                 `json:"method"`
	ClientToken string                 `json:"clientToken"`
	Params      map[string]interface{} `json:"params"`
	Timestamp   int64                  `json:"timestamp,omitempty"`

	Type     string `json:"type,omitempty"`
	ActionId string `json:"actionId,omitempty"`
}

func (p *ExplorerPayload) GetMethodOrType() string {
	if p.Method != "" {
		return p.Method
	}
	if p.Type != "" {
		return p.Type
	}

	logrus.Warnf("not found methd and type. [method=%v][type=%v]", p.Method, p.Type)
	return ""
}

func (resp *DeviceCtx) subAllTopics() error {
	subcribedTopics := []string{
		// ⬇️ 物模型 topic
		fmt.Sprintf("$thing/down/property/%s/%s", resp.ProductId, resp.DeviceName), // 物模型属性
		fmt.Sprintf("$thing/down/service/%s/%s", resp.ProductId, resp.DeviceName),  // 物模型服务
		fmt.Sprintf("$thing/down/action/%s/%s", resp.ProductId, resp.DeviceName),   // 物模型行为
		fmt.Sprintf("$thing/down/raw/%s/%s", resp.ProductId, resp.DeviceName),      // 二进制
		fmt.Sprintf("$thing/down/event/%s/%s", resp.ProductId, resp.DeviceName),    // 事件

		// ⬇️ 系统 Topic
		fmt.Sprintf("$ota/update/%s/%s", resp.ProductId, resp.DeviceName),              // OTA
		fmt.Sprintf("$broadcast/rxd/%s/%s", resp.ProductId, resp.DeviceName),           // 广播消息
		fmt.Sprintf("$shadow/operation/result/%s/%s", resp.ProductId, resp.DeviceName), // 设备影子
		fmt.Sprintf("$rrpc/rxd/%s/%s/+", resp.ProductId, resp.DeviceName),              // RRPC
		fmt.Sprintf("$sys/operation/result/%s/%s/+", resp.ProductId, resp.DeviceName),  // ntp

		// ⬇️ 自定义 topic
		fmt.Sprintf("%s/%s/data", resp.ProductId, resp.DeviceName),
		fmt.Sprintf("%s/%s/control", resp.ProductId, resp.DeviceName),
	}
	var eg errgroup.Group
	for _, topic := range subcribedTopics {
		cpTopic := topic
		eg.Go(func() error {
			return resp.subTopic(cpTopic)
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

func (resp *DeviceCtx) subTopic(topic string) (err error) {
	token := resp.mqttClient.Subscribe(topic, 1, resp.subHandler)
	if wa, err := token.Wait(), token.Error(); !wa || err != nil {
		return errors.Cause(err)
	}

	return nil
}

func (resp *DeviceCtx) Publish(topic string, payload []byte) {
	client := resp.mqttClient
	token := client.Publish(topic, 1, false, payload)
	clog.Logger(resp.ctx).WithField("topic", topic).Debugf("发送消息:\n%s\n", utils.GetPrettyJSONStr(string(payload)))
	if notTimeout := token.WaitTimeout(time.Second * 2); !notTimeout {
		clog.Logger(resp.ctx).Warnf("发送消息超时.")
	}
	if token.Error() != nil {
		clog.Logger(resp.ctx).WithError(token.Error()).Errorf("发送消息失败")
	}
}
