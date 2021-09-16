package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"testUtils/fakeDevice/clog"
	"testUtils/fakeDevice/config"
	"testUtils/fakeDevice/device"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

var (
	cfg = pflag.StringP("config", "c", "", "config file path.")
)

func init() {

	pflag.Parse()
	logrus.SetLevel(logrus.TraceLevel)
	if err := config.InitViper(cfg); err != nil {
		clog.Logger().WithError(err).Panic(err)
	}
}
func main() {

	devices, err := config.GetDevices()
	if err != nil {
		logrus.Errorf("%+v", err)
		return
	}
	devicesCtx := make([]device.DeviceInterface, 0, len(devices))
	for _, v := range devices {
		deviceCtx, err := device.GetDeviceCtx(context.Background(), v, config.GetURLSuff())
		if err != nil {
			logrus.Errorf("err:%v", err)
		}
		devicesCtx = append(devicesCtx, deviceCtx)
		go deviceCtx.Connect()
	}

	// 主线程阻塞
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs
	logrus.Infof("收到信号: %v", sig)
	for _, v := range devicesCtx {
		v.Disconnect()
	}
	time.Sleep(time.Millisecond * 200)
	logrus.Warnf("退出")
	return

}
