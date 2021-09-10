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

func (resp *DeviceCtx) ReportOTAVersion(version string) {
	type reportOTAVersion struct {
		Type   string `json:"type"`
		Report struct {
			Version string `json:"version"`
		} `json:"report"`
	}
	sendPayload := reportOTAVersion{
		Type: "report_version",
		Report: struct {
			Version string "json:\"version\""
		}{
			Version: version,
		},
	}
	stBytes, _ := json.Marshal(sendPayload)
	topic := fmt.Sprintf("$ota/report/%s/%s", resp.ProductId, resp.DeviceName)
	logrus.Infof("report ota version")
	publish(resp.MQTTClient, topic, stBytes)
}
