package main

import (
	"os"
	"os/signal"
	"syscall"
	"testUtils/fakeDevice/config"
	"testUtils/fakeDevice/device"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	cfg = pflag.StringP("config", "c", "", "config file path.")
)

func init() {

	pflag.Parse()
	logrus.SetLevel(logrus.TraceLevel)
	if err := config.InitViper(cfg); err != nil {
		logrus.WithError(err).Panic("初始化配置文件失败")
	}
}
func main() {

	defer viper.WriteConfigAs("./newconfig_back.yaml")
	devices, err := config.GetDevices()
	if err != nil {
		logrus.Errorf("%+v", err)
		return
	}
	URLSuff := config.GetURLSuff()
	for _, v := range devices {
		deviceCtx, err := device.GetDeviceCtx(v, URLSuff)
		if err != nil {
			logrus.Errorf("err:%v", err)
		}
		go deviceCtx.Connect()
		defer deviceCtx.Disconnect()
	}

	// 主线程阻塞
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs
	logrus.Infof("receive : %v", sig)
	logrus.Warnf("exiting")
	return

}
