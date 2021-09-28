package device

import (
	"encoding/json"
	"fmt"
	"testUtils/fakeDevice/clog"
	"testUtils/fakeDevice/device/shadow"

	"github.com/google/uuid"
)

type shadowPayload struct {
	ClientToken string `json:"clientToken"`
	Payload     shadow.ShadowInfo
	Result      int    `json:"result"`
	Timestamp   int    `json:"timestamp"`
	Type        string `json:"type"`
}

func (resp *DeviceCtx) OnShadowDown(payload shadowPayload) {
	if payload.Result != 0 {
		clog.Logger(resp.ctx).Warnf("影子处理错误，错误码: %d", payload.Result)
	}
	resp.shadow = payload.Payload
	if resp.shadow.State.Desired != nil {
		resp.ShadowReport(payload.Payload)
	}
}

func (resp *DeviceCtx) ShadowGet() {

	sendPayload := shadow.ShadowUploadReq{
		Type:        "get",
		ClientToken: uuid.New().String(),
	}
	stBytes, _ := json.Marshal(sendPayload)
	topic := fmt.Sprintf("$shadow/operation/%s/%s", resp.ProductId, resp.DeviceName)
	clog.Logger(resp.ctx).Infof("获取影子状态(shadow get)")
	resp.Publish(topic, stBytes)
}

func (resp *DeviceCtx) ShadowReport(reportedShadow shadow.ShadowInfo) {
	shadowInfoReported := shadow.ShadowUploadReq{
		Type:        "update",
		Version:     reportedShadow.Version,
		ClientToken: uuid.NewString(),
	}
	shadowInfoReported.State.Reported = reportedShadow.State.Desired
	jsonBytes, err := json.Marshal(shadowInfoReported)
	if err != nil {
		clog.Logger(resp.ctx).Warnf("err:%v", err)
	}
	topic := fmt.Sprintf("$shadow/operation/%s/%s", resp.ProductId, resp.DeviceName)
	clog.Logger(resp.ctx).Infof("上报影子")
	resp.Publish(topic, jsonBytes)

}
