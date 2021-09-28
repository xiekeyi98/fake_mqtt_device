package config

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/wire"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var Provider = wire.NewSet(GetDevices)

func init() {
	viper.SetDefault("URLSuff", "iotcloud.tencentdevices.com")
}

// 设备配置信息
type Device struct {
	ProductId     string
	DeviceName    string
	Psk           string
	MQTTHost      string
	DeviceVersion string

	Protocol string
}

func (d *Device) IsShadow() bool {
	if strings.Contains(d.Protocol, "shadow") {
		return true
	}
	return false
}

func (d *Device) GetMQTTDSN() (*MQTTDSN, error) {
	rd := rand.Int31n(100)
	res := &MQTTDSN{
		Broker:   fmt.Sprintf("%s.%s", d.ProductId, d.GetURLSuff()),
		Port:     1883,
		UserName: fmt.Sprintf("%s%s;%s;%d;%d", d.ProductId, d.DeviceName, "12010126", rd, time.Now().Unix()+24*3600),
	}
	if d.MQTTHost != "" {
		res.Broker = d.MQTTHost
	}
	token, err := res.GetDeviceToken(d.Psk)
	if err != nil {
		return nil, err
	}
	res.Token = token
	return res, nil
}

func (m *MQTTDSN) GetDeviceToken(deicePsk string) (string, error) {
	if m.Token != "" {
		return m.Token, nil
	}
	rawKey, err := base64.StdEncoding.DecodeString(deicePsk)
	if err != nil {
		return "", errors.Cause(err)
	}
	h := hmac.New(sha1.New, rawKey)
	_, err = h.Write([]byte(m.UserName))
	if err != nil {
		return "", errors.Cause(err)
	}
	token := hex.EncodeToString(h.Sum(nil))
	token = fmt.Sprintf("%s;%s", token, "hmacsha1")
	m.Token = token
	return token, nil

}

type MQTTDSN struct {
	Broker   string
	Port     int
	UserName string
	Token    string
}

func GetDevices() ([]Device, error) {
	devices := []Device{}
	if err := viper.UnmarshalKey("Devices", &devices); err != nil {
		return devices, errors.Cause(err)
	}
	return devices, nil
}

func GetDevice(ProductId, Devicename string) (*Device, error) {
	devices, _ := GetDevices() // performance
	for _, device := range devices {
		if device.ProductId == ProductId && device.DeviceName == Devicename {
			return &device, nil
		}
	}
	return nil, errors.New("not found.")

}
func (d *Device) SetDeviceVersion(version string) {
	logrus.Infof("set device to new version.")
	d.DeviceVersion = version
}

func (d *Device) GetURLSuff() string {
	return viper.GetString("URLSuff")
}

type Actions struct {
	ActionId string
	Params   ActionParam
}
type ActionParam map[string]interface{}

func GetActionParams(actionId string) (Param ActionParam, err error) {

	Actions := make([]Actions, 0)
	if err := viper.UnmarshalKey("Actions", &Actions); err != nil {
		return Param, errors.Cause(err)
	}
	for _, v := range Actions {
		if v.ActionId == actionId {
			return v.Params, nil
		}
	}
	logrus.Warnf("actionId not found, use nil.")
	return Param, nil
}

type Event struct {
	SendInterval time.Duration
	EventInfos   []EventInfo
}

type EventInfo struct {
	EventId     string
	EventType   string
	EventParams map[string]interface{}
}

func GetEvent() (Event, error) {
	event := Event{}
	if err := viper.UnmarshalKey("Events", &event); err != nil {
		return event, errors.Cause(err)
	}
	return event, nil
}

type OTAConf struct {
	DownloadingTime time.Duration
	BurningTime     time.Duration
}

func GetOTAConf() (OTAConf, error) {
	conf := OTAConf{}
	if err := viper.UnmarshalKey("OTA", &conf); err != nil {
		return conf, errors.Cause(err)
	}
	return conf, nil
}
