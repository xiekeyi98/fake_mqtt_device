package device

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func (resp *DeviceCtx) GetStatus() {

	type getStatusStruct struct {
		Method      string `json:"method"`
		ClientToken string `json:"clientToken"`
		Type        string `json:"type"`
		ShowMeta    int    `json:"showmeta"`
	}

	sendPayload := getStatusStruct{
		Method:      "get_status",
		ClientToken: uuid.New().String(),
		Type:        "report",
		ShowMeta:    1,
	}
	stBytes, _ := json.Marshal(sendPayload)
	topic := fmt.Sprintf("$thing/up/property/%s/%s", resp.ProductId, resp.DeviceName)
	logrus.Infof("get device status")
	publish(resp.MQTTClient, topic, stBytes)
}
