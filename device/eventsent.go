package device

import (
	"encoding/json"
	"fmt"
	"testUtils/fakeDevice/clog"
	"testUtils/fakeDevice/config"
	"time"

	"github.com/google/uuid"
)

func (resp *DeviceCtx) ReportEvents() {

	event, err := config.GetEvent()
	if err != nil {
		clog.Logger(resp.ctx).WithError(err).Errorf("获取事件配置错误")
		return
	}
	if event.SendInterval < time.Millisecond {
		clog.Logger(resp.ctx).Infof("间隔过短或关闭事件上报，跳过事件上报")
		return
	}
	tic := time.NewTicker(event.SendInterval)
	clog.Logger(resp.ctx).Infof("开始每 %v 时间上报一次事件", event.SendInterval)
	defer tic.Stop()
	for {
		select {
		case <-tic.C:
			for _, v := range event.EventInfos {
				resp.reportEvent(v)
			}
		case <-resp.ctx.Done():
			clog.Logger(resp.ctx).Infof("事件管理因 context 取消结束。")
			return

		}
	}
}

func (resp *DeviceCtx) reportEvent(eventInfo config.EventInfo) {
	type sendPayloadStruct struct {
		Method      string                 `json:"method"`
		ClientToken string                 `json:"clientToken"`
		Version     string                 `json:"version"`
		EventId     string                 `json:"eventId"`
		Type        string                 `json:"type"`
		Timestamp   int64                  `json:"timestamp"`
		Params      map[string]interface{} `json:"params"`
	}

	sendPayload := sendPayloadStruct{
		Method:      "event_post",
		ClientToken: uuid.New().String(),
		Version:     "1.0",
		EventId:     eventInfo.EventId,
		Type:        eventInfo.EventType,
		Timestamp:   time.Now().UnixNano() / int64(time.Millisecond),
		Params:      eventInfo.EventParams,
	}
	stBytes, _ := json.Marshal(sendPayload)
	topic := fmt.Sprintf("$thing/up/event/%s/%s", resp.ProductId, resp.DeviceName)
	resp.Publish(topic, stBytes)
}
