package ota

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testUtils/fakeDevice/clog"
	"testUtils/fakeDevice/config"
	mqttClient "testUtils/fakeDevice/device/mqttInterface"
	"testUtils/fakeDevice/utils"
	"time"

	"github.com/sirupsen/logrus"
)

type OTATaskStruct struct {
	ctx context.Context
	mqttClient.Interface
	ProductId  string
	DeviceName string
}

type updateFirmware struct {
	Filesize int    `json:"file_size"`
	Md5sum   string `json:"md5sum"`
	Type     string `json:"update_firmware"`
	Url      string `json:"url"`
	Version  string `json:"version"`
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

func NewOTATask(ctx context.Context, ProductId, DeviceName string, Inter mqttClient.Interface) *OTATaskStruct {
	return &OTATaskStruct{
		ctx:        ctx,
		Interface:  Inter,
		ProductId:  ProductId,
		DeviceName: DeviceName,
	}
}

func (o *OTATaskStruct) OTAFirmwareUpdate(RawFirmwareUpdate string) {

	firmware := updateFirmware{}
	if err := json.Unmarshal([]byte(RawFirmwareUpdate), &firmware); err != nil {
		clog.Logger(o.ctx).WithError(err).Errorf("固件升级信息解析失败")
		return
	}
	httpClient := http.DefaultClient
	httpClient.Timeout = time.Second * 3
	firmwareResp, err := httpClient.Get(firmware.Url)
	if err != nil {
		clog.Logger(o.ctx).WithError(err).Errorf("下载固件失败.")
		return
	}
	defer firmwareResp.Body.Close()
	firmwareBodys, err := io.ReadAll(firmwareResp.Body)
	if err != nil {
		clog.Logger(o.ctx).WithError(err).Errorf("下载固件失败.")
	}
	md5Calc := utils.GetMd5(string(firmwareBodys))
	clog.Logger(o.ctx).Infof("获取到大小为 %v 字节，md5 为: %v 的固件(%v)", len(firmwareBodys), string(md5Calc), firmware.Version)
	if md5Calc != firmware.Md5sum {
		clog.Logger(o.ctx).Warnf("md5 不符合。服务器md5:%s,客户端计算:%s", firmware.Md5sum, md5Calc)
	}
	if len(firmwareBodys) != firmware.Filesize {
		clog.Logger(o.ctx).Warnf("固件大小不符合。服务器大小:%d,客户端计算:%d", firmware.Filesize, len(firmwareBodys))
	}

	OTAConf, err := config.GetOTAConf()
	if err != nil {
		logrus.Errorf("Err:%v", err)
		return
	}
	o.ReportProgress("downloading", firmware.Version, OTAConf.DownloadingTime)
	time.Sleep(time.Second)
	o.ReportProgress("burning", firmware.Version, OTAConf.BurningTime)
	time.Sleep(time.Second)
	o.ReportSucc(firmware.Version)
}

func (o *OTATaskStruct) ReportProgress(ProgressType, targetVersion string, ReportTime time.Duration) {

	clog.Logger(o.ctx).Infof("开始上报固件升级 %v 进度", ProgressType)
	if ReportTime < time.Second {
		ReportTime = time.Second
	}
	tic := time.NewTicker(time.Second * 2)
	defer tic.Stop()
	percent := 0
	progressReport := otaReport{}
	progressReport.Type = "report_progress"
	progressReport.Report.Progress.State = ProgressType
	progressReport.Report.Progress.Percent = fmt.Sprintf("%d", percent)
	progressReport.Report.Progress.ResultCode = "0"
	progressReport.Report.Progress.ResultMsg = "succ"
	progressReport.Report.Version = targetVersion
forLabel:
	for {
		select {
		case <-tic.C:
			percent = percent + int(100/ReportTime.Seconds())
			if percent > 100 {
				percent = 100
			}
			clog.Logger(o.ctx).Debugf("上报 %v 进度:%d", ProgressType, percent)
			progressReport.Report.Progress.Percent = fmt.Sprintf("%d", percent)
			stBytes, _ := json.Marshal(progressReport)
			o.Publish(fmt.Sprintf("$ota/report/%s/%s", o.ProductId, o.DeviceName), stBytes)
			if percent >= 100 {
				break forLabel
			}
		case <-o.ctx.Done():
			clog.Logger(o.ctx).Warnf("上报 %v 进度取消。", ProgressType)
			o.ReportFailed(targetVersion, fmt.Sprintf("%s %v", ProgressType, o.ctx.Err()))
			return
		}
	}
}

func (o *OTATaskStruct) ReportSucc(targetVersion string) {
	report := otaReport{}
	report.Type = "report_progress"
	report.Report.Progress.State = "done"
	report.Report.Progress.ResultMsg = "succ"
	report.Report.Progress.ResultCode = "0"
	report.Report.Version = targetVersion
	stBytes, _ := json.Marshal(report)
	o.Publish(fmt.Sprintf("$ota/report/%s/%s", o.ProductId, o.DeviceName), stBytes)
	clog.Logger(o.ctx).Infof("OTA 升级结束。")
}

func (o *OTATaskStruct) ReportFailed(targetVersion string, errMessage string) {
	report := otaReport{}
	report.Type = "report_progress"
	report.Report.Progress.State = "fail"
	report.Report.Progress.ResultMsg = errMessage
	report.Report.Progress.ResultCode = "-5"
	report.Report.Version = targetVersion
	stBytes, _ := json.Marshal(report)
	o.Publish(fmt.Sprintf("$ota/report/%s/%s", o.ProductId, o.DeviceName), stBytes)
	clog.Logger(o.ctx).Infof("OTA 升级以失败结束。")
}
