package device

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/rand"
	"testUtils/fakeDevice/clog"
	"testUtils/fakeDevice/config"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

type DeviceInterface interface {
	Connect() error
	Disconnect()
}

type DeviceCtx struct {
	MQTTClient mqtt.Client

	config.Device
	ctx       context.Context
	cancelCtx context.CancelFunc

	broker   string
	port     int
	userName string
	token    string
}

// 生成设备密钥
func generateDeviceToken(userName, psk string) (string, error) {
	rawKey, err := base64.StdEncoding.DecodeString(psk)
	if err != nil {
		return "", errors.Cause(err)
	}
	h := hmac.New(sha1.New, rawKey)
	_, err = h.Write([]byte(userName))
	if err != nil {
		return "", errors.Cause(err)
	}
	token := hex.EncodeToString(h.Sum(nil))
	token = fmt.Sprintf("%s;%s", token, "hmacsha1")
	return token, nil
}

// 获取一个设备抽象
func GetDeviceCtx(ctx context.Context, d config.Device, URLSuff string) (DeviceInterface, error) {

	ctx, cancel := context.WithCancel(ctx)
	ctx = clog.WithCtx(ctx, "ProductId", d.ProductId)
	ctx = clog.WithCtx(ctx, "DeviceName", d.DeviceName)
	rd := rand.Int31n(100)
	resp := DeviceCtx{
		ctx:       ctx,
		cancelCtx: cancel,
		Device:    d,
		broker:    fmt.Sprintf("%s.%s", d.ProductId, URLSuff),
		port:      1883,
		userName:  fmt.Sprintf("%s%s;%s;%d;%d", d.ProductId, d.DeviceName, "12010126", rd, time.Now().Unix()+10*365*24*3600),
	}
	if d.MQTTHost != "" {
		resp.broker = d.MQTTHost
	}
	if token, err := generateDeviceToken(resp.userName, resp.Psk); err != nil {
		return nil, err
	} else {
		resp.token = token
	}

	mqttClientOpts := mqtt.
		NewClientOptions().
		AddBroker(fmt.Sprintf("tcp://%s:%d", resp.broker, resp.port)).
		SetClientID(fmt.Sprintf("%s%s", d.ProductId, d.DeviceName)).
		SetUsername(resp.userName).
		SetPassword(resp.token)

	mqttClientOpts.OnConnectionLost = func(client mqtt.Client, err error) {
		clog.Logger(resp.ctx).WithError(err).Errorf("connect lost.")
	}
	client := mqtt.NewClient(mqttClientOpts)
	resp.MQTTClient = client

	return &resp, nil
}
func (resp *DeviceCtx) Connect() error {
	token := resp.MQTTClient.Connect()
	if wa, err := token.Wait(), token.Error(); !wa || err != nil {
		clog.Logger(resp.ctx).WithError(err).Errorf("等待 MQTT 完成错误")
		return errors.Cause(err)
	}
	if err := resp.subAllTopics(); err != nil {
		clog.Logger(resp.ctx).WithError(err).Errorf("订阅 topic 失败")
		return errors.Cause(err)
	}
	clog.Logger(resp.ctx).Infof("连接成功")
	resp.postConnect()
	return nil
}

func (resp *DeviceCtx) postConnect() {
	resp.GetStatus()
	resp.ReportOTAVersion(resp.Device.DeviceVersion)
	go resp.ReportEvents()

}

func (resp *DeviceCtx) Disconnect() {
	resp.cancelCtx()
	resp.MQTTClient.Disconnect(2000)
}

type Payload struct {
	Method      string                 `json:"method"`
	ClientToken string                 `json:"clientToken"`
	Params      map[string]interface{} `json:"params"`
	Timestamp   int64                  `json:"timestamp,omitempty"`
	Type        string                 `json:"type"`

	ActionId string `json:"actionId,omitempty"`
}

func (p *Payload) GetMethodOrType() string {
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
		fmt.Sprintf("$thing/down/property/%s/%s", resp.ProductId, resp.DeviceName), // 物模型属性
		fmt.Sprintf("$thing/down/service/%s/%s", resp.ProductId, resp.DeviceName),  // 物模型服务
		fmt.Sprintf("$thing/down/action/%s/%s", resp.ProductId, resp.DeviceName),   // 物模型行为
		fmt.Sprintf("$thing/down/raw/%s/%s", resp.ProductId, resp.DeviceName),      // 二进制
		fmt.Sprintf("$thing/down/event/%s/%s", resp.ProductId, resp.DeviceName),    // 事件
		fmt.Sprintf("$ota/update/%s/%s", resp.ProductId, resp.DeviceName),          // OTA
	}
	for _, topic := range subcribedTopics {
		if err := resp.subTopic(topic); err != nil {
			return err
		}
	}
	return nil
}

func (resp *DeviceCtx) subTopic(topic string) (err error) {
	token := resp.MQTTClient.Subscribe(topic, 1, resp.subHandler)
	if wa, err := token.Wait(), token.Error(); !wa || err != nil {
		return errors.Cause(err)
	}

	return nil
}

func (resp *DeviceCtx) publish(topic string, payload interface{}) {
	client := resp.MQTTClient
	token := client.Publish(topic, 1, false, payload)
	clog.Logger(resp.ctx).WithField("topic", topic).Debugf("发送消息:%s", cast.ToString(payload))
	token.Wait()
	if token.Error() != nil {
		clog.Logger(resp.ctx).WithError(token.Error()).Errorf("发送消息失败")
	}
}
