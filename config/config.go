package config

import (
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault("URLSuff", "iotcloud.tencentdevices.com")
}

type Device struct {
	ProductId     string
	DeviceName    string
	Psk           string
	MQTTHost      string
	DeviceVersion string
}

func GetDevices() ([]Device, error) {
	devices := []Device{}
	if err := viper.UnmarshalKey("Devices", &devices); err != nil {
		return devices, errors.Cause(err)
	}
	return devices, nil
}

func GetDevice(ProductId, Devicename string) (*Device, error) {
	devices, _ := GetDevices()
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

func GetURLSuff() string {
	URLSuff := viper.GetString("URLSuff")
	return URLSuff
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
