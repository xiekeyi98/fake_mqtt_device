package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"testUtils/fakeDevice/clog"
	"testUtils/fakeDevice/config"
	"testUtils/fakeDevice/device"
	"testUtils/fakeDevice/utils"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

var (
	cfg     = pflag.StringP("config", "c", "", "config file path.")
	verbose = pflag.BoolP("verbose", "v", false, "verbose output")
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})
	logrus.SetOutput(colorable.NewColorableStdout())
	clog.Logger().Infof("Powered by keyixie.[see at github:https://github.com/xiekeyi98/fake_mqtt_device]")
	pflag.Parse()
	logrus.SetLevel(logrus.InfoLevel)
	if verbose != nil && *verbose {
		logrus.Infof("use verbose mode.")
		logrus.SetLevel(logrus.TraceLevel)

	}
	if err := config.InitViper(cfg); err != nil {
		clog.Logger().WithError(err).Panic(err)
	}
	clog.Logger().Debugf("完整配置：%s", utils.GetPrettyJSON(viper.AllSettings()))
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
			clog.Logger().Errorf("err:%+v", err)
		}
		devicesCtx = append(devicesCtx, deviceCtx)
		errG.Go(deviceCtx.Connect)
	}
	if err := errG.Wait(); err != nil {
		clog.Logger().WithError(err).Errorf("建立设备连接失败:%+v", err)
	}

	// 主线程阻塞
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs
	clog.Logger().Infof("收到信号: %v", sig)
	for _, v := range devicesCtx {
		go v.Disconnect()
	}
	time.Sleep(time.Millisecond * 300)
	clog.Logger().Warnf("退出")
	return

}
