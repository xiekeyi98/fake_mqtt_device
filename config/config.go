package config

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault("URLSuff", "iotcloud.tencentdevices.com")
}

type Device struct {
	ProductId  string
	DeviceName string
	Psk        string
	MQTTHost   string
}

func GetDevices() ([]Device, error) {
	devices := []Device{}
	if err := viper.UnmarshalKey("Devices", &devices); err != nil {
		return devices, errors.Cause(err)
	}
	return devices,nil
}

func GetURLSuff() string {
	URLSuff := viper.GetString("URLSuff")
	return URLSuff
}
