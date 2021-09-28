package device

import (
	"encoding/json"
	"fmt"
	"testUtils/fakeDevice/clog"

	"github.com/google/uuid"
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
	clog.Logger(resp.ctx).Infof("获取设备状态(get_status)")
	resp.Publish(topic, stBytes)
}

func (resp *DeviceCtx) ReportOTAVersion(version string) {
	if version == "" {
		clog.Logger(resp.ctx).Infof("设备版本为空，跳过版本上报")
		return
	}
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
	clog.Logger(resp.ctx).Infof("上报 OTA 版本 %s", version)
	resp.Publish(topic, stBytes)
}
