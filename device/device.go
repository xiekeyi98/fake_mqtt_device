package device

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/rand"
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

	broker   string
	port     int
	userName string
	token    string
}

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
func GetDeviceCtx(d config.Device, URLSuff string) (DeviceInterface, error) {
	logger := logrus.WithFields(
		map[string]interface{}{
			"product_id":  d.ProductId,
			"device_name": d.DeviceName,
		},
	)
	rd := rand.Int31n(100)
	resp := DeviceCtx{
		Device:   d,
		broker:   fmt.Sprintf("%s.%s", d.ProductId, URLSuff),
		port:     1883,
		userName: fmt.Sprintf("%s%s;%s;%d;%d", d.ProductId, d.DeviceName, "12010126", rd, time.Now().Unix()+10*365*24*3600),
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
		//SetDefaultPublishHandler(messagePubHandler)
	mqttClientOpts.OnConnect = func(client mqtt.Client) {
	}
	mqttClientOpts.OnConnectionLost = func(client mqtt.Client, err error) {
		logger.Errorf("Connect lost: %v", err)
	}
	client := mqtt.NewClient(mqttClientOpts)
	resp.MQTTClient = client

	return &resp, nil
}
func (resp *DeviceCtx) Connect() error {
	token := resp.MQTTClient.Connect()
	if wa, err := token.Wait(), token.Error(); !wa || err != nil {
		logrus.Errorf("wait err:%v", err)
		return err
	}
	if err := resp.subAllTopics(); err != nil {
		logrus.Errorf("sub all topics err :%v", err)
		return err
	}
	logrus.Infof("%s/%s@%s connect succ", resp.ProductId, resp.DeviceName, resp.broker)
	resp.GetStatus()
	go resp.ReportEvents()
	return nil
}

func (resp *DeviceCtx) Disconnect() {
	resp.MQTTClient.Disconnect(2000)
}

type Payload struct {
	Method      string                 `json:"method"`
	ClientToken string                 `json:"clientToken"`
	Params      map[string]interface{} `json:"params"`
	Timestamp   int64                  `json:"timestamp,omitempty"`

	ActionId string `json:"actionId,omitempty"`
}

func (resp *DeviceCtx) subAllTopics() error {
	subcribedTopics := []string{
		fmt.Sprintf("$thing/down/property/%s/%s", resp.ProductId, resp.DeviceName),
		fmt.Sprintf("$thing/down/service/%s/%s", resp.ProductId, resp.DeviceName),
		fmt.Sprintf("$thing/down/action/%s/%s", resp.ProductId, resp.DeviceName),
		fmt.Sprintf("$thing/down/raw/%s/%s", resp.ProductId, resp.DeviceName),
		fmt.Sprintf("$thing/down/event/%s/%s", resp.ProductId, resp.DeviceName),
	}
	for _, topic := range subcribedTopics {
		if err := resp.subTopic(topic); err != nil {
			return err
		}
	}
	return nil
}

func (resp *DeviceCtx) subTopic(topic string) (err error) {
	logger := logrus.WithFields(
		map[string]interface{}{
			"ProductId":  resp.ProductId,
			"DeviceName": resp.Device,
		},
	)
	token := resp.MQTTClient.Subscribe(topic, 1, resp.subHandler)
	if wa, err := token.Wait(), token.Error(); !wa || err != nil {
		logger.Errorf("token.Wait:%v,err:%+v", wa, err)
		return errors.Cause(err)
	}

	return nil
}

func publish(client mqtt.Client, topic string, payload interface{}) {
	token := client.Publish(topic, 1, false, payload)
	logrus.Debugf("publish response to %v.paylord:%v", topic, cast.ToString(payload))
	token.Wait()
	if token.Error() != nil {
		logrus.Warnf("err:%v", token.Error())
	}
}
