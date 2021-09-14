package device

import (
	"encoding/json"
	"fmt"
	"testUtils/fakeDevice/config"
	"time"

	"github.com/sirupsen/logrus"
)

func (resp *DeviceCtx) OTAReport(rawmessage string) {
	type updateFirmware struct {
		Filesize int    `json:"file_size"`
		Md5sum   string `json:"md5sum"`
		Type     string `json:"update_firmware"`
		Url      string `json:"url"`
		Version  string `json:"version"`
	}
	firmware := updateFirmware{}
	if err := json.Unmarshal([]byte(rawmessage), &firmware); err != nil {
		logrus.Errorf("Err:%v", err)
		return
	}
	OTAConf, err := config.GetOTAConf()
	if err != nil {
		logrus.Errorf("Err:%v", err)
		return
	}
	resp.reportDownloading(firmware.Version, OTAConf.DownloadingTime)
	time.Sleep(time.Second)
	resp.reportBurning(firmware.Version, OTAConf.BurningTime)
	time.Sleep(time.Second)
	resp.reportDone(firmware.Version)
	device, err := config.GetDevice(resp.ProductId, resp.DeviceName)
	if err != nil {
		logrus.Errorf("err:+%v", err)
		return
	}
	device.SetDeviceVersion(firmware.Version)
	resp.ReportOTAVersion(firmware.Version)
}

type otaReport struct {
	Type   string `json:"type"`
	Report struct {
		Version  string `json:"version"`
		Progress struct {
			State      string `json:"state"`
			Percent    string `json:"percent,omitempty"`
			ResultCode string `json:"result_code"`
			ResultMsg  string `json:"result_msg"`
		} `json:"progress"`
	} `json:"report"`
}

func (resp *DeviceCtx) reportDownloading(targetVersion string, DownloadingTime time.Duration) {

	logrus.Infof("start to report download progress.")
	if DownloadingTime < time.Second {
		DownloadingTime = time.Second
	}
	tic := time.NewTicker(DownloadingTime / 20)
	defer tic.Stop()
	percent := 0
	for range tic.C {
		progressReport := otaReport{}
		progressReport.Type = "report_progress"
		progressReport.Report.Progress.State = "downloading"
		progressReport.Report.Progress.Percent = fmt.Sprintf("%d", percent)
		progressReport.Report.Progress.ResultCode = "0"
		progressReport.Report.Progress.ResultMsg = "succ"
		progressReport.Report.Version = targetVersion
		logrus.Debugf("report downloading percent :%d", percent)
		stBytes, _ := json.Marshal(progressReport)
		publish(resp.MQTTClient, fmt.Sprintf("$ota/report/%s/%s", resp.ProductId, resp.DeviceName), stBytes)
		percent += 5
		if percent > 100 {
			break
		}
	}
	return
}

func (resp *DeviceCtx) reportBurning(targetVersion string, buringTime time.Duration) {

	logrus.Infof("start to report buring.")
	if buringTime < time.Second {
		buringTime = time.Second
	}
	tic := time.NewTicker(buringTime / 20)
	defer tic.Stop()
	percent := 0
	for range tic.C {
		progressReport := otaReport{}
		progressReport.Type = "report_progress"
		progressReport.Report.Progress.State = "burning"
		progressReport.Report.Progress.Percent = fmt.Sprintf("%d", percent)
		progressReport.Report.Progress.ResultCode = "0"
		progressReport.Report.Progress.ResultMsg = "succ"
		progressReport.Report.Version = targetVersion
		stBytes, _ := json.Marshal(progressReport)
		logrus.Debugf("report buring percent :%d", percent)
		publish(resp.MQTTClient, fmt.Sprintf("$ota/report/%s/%s", resp.ProductId, resp.DeviceName), stBytes)
		percent += 5
		if percent > 100 {
			break
		}
	}
	return
}

func (resp *DeviceCtx) reportDone(targetVersion string) {

	logrus.Infof("report ota done.")
	progressReport := otaReport{}
	progressReport.Type = "report_progress"
	progressReport.Report.Progress.State = "done"
	progressReport.Report.Progress.ResultMsg = "succ"
	progressReport.Report.Progress.ResultCode = "0"
	progressReport.Report.Version = targetVersion
	stBytes, _ := json.Marshal(progressReport)
	publish(resp.MQTTClient, fmt.Sprintf("$ota/report/%s/%s", resp.ProductId, resp.DeviceName), stBytes)
}
