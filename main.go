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

	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"
)

var (
	cfg = pflag.StringP("config", "c", "", "config file path.")
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})
	logrus.SetOutput(colorable.NewColorableStdout())
	logrus.Infof("Powered by keyixie.[see at github:https://github.com/xiekeyi98/fake_mqtt_device]")
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
	var errG errgroup.Group
	for _, v := range devices {
		deviceCtx, err := device.GetDeviceCtx(context.Background(), v)
		if err != nil {
			logrus.Errorf("err:%+v", err)
		}
		devicesCtx = append(devicesCtx, deviceCtx)
		errG.Go(deviceCtx.Connect)
	}
	if err := errG.Wait(); err != nil {
		logrus.Errorf("建立设备连接失败:%+v", err)
	}

	// 主线程阻塞
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs
	logrus.Infof("收到信号: %v", sig)
	for _, v := range devicesCtx {
		go v.Disconnect()
	}
	time.Sleep(time.Millisecond * 300)
	logrus.Warnf("退出")
	return

}
