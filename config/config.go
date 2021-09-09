package config

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

type Actions struct {
	ActionId string
	Params   ActionParam
}
type ActionParam map[string]interface{}

func GetDevices() ([]Device, error) {
	devices := []Device{}
	if err := viper.UnmarshalKey("Devices", &devices); err != nil {
		return devices, errors.Cause(err)
	}
	return devices, nil
}

func GetURLSuff() string {
	URLSuff := viper.GetString("URLSuff")
	return URLSuff
}

func GetActionParams(actionId string) (Param ActionParam, err error) {

	Actions := make([]Actions, 0, 0)
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
