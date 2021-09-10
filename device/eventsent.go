package device

import (
	"encoding/json"
	"fmt"
	"testUtils/fakeDevice/config"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func (resp *DeviceCtx) ReportEvents() {

	event, err := config.GetEvent()
	if err != nil {
		logrus.Errorf("get events error:%+v", err)
	}
	if event.SendInterval <= time.Millisecond {
		logrus.Infof("no event need to be sent.")
		return
	}
	tic := time.NewTicker(event.SendInterval)
	logrus.Infof("start to send event every %v", event.SendInterval)
	defer tic.Stop()
	for range tic.C {
		for _, v := range event.EventInfos {
			resp.reportEvent(v)

		}
	}
}

func (resp *DeviceCtx) reportEvent(eventInfo config.EventInfo) {
	type sendPayloadStruct struct {
		Method      string `json:"method"`
		ClientToken string `json:"clientToken"`
		//Version     string `json:"version"`
		EventId string `json:"eventId"`
		//Type        string                 `json:"type"`
		Timestamp int64                  `json:"timestamp"`
		Params    map[string]interface{} `json:"params"`
	}

	sendPayload := sendPayloadStruct{
		Method:      "event_post",
		ClientToken: uuid.New().String(),
		//Version: "1.0",
		EventId: eventInfo.EventId,
		//Type:      eventInfo.EventType,
		Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
		Params:    eventInfo.EventParams,
	}
	stBytes, _ := json.Marshal(sendPayload)
	topic := fmt.Sprintf("$thing/up/event/%s/%s", resp.ProductId, resp.DeviceName)
	publish(resp.MQTTClient, topic, stBytes)
}
