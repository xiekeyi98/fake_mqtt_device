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
	if cfg != nil && *cfg != "" {
		viper.SetConfigFile(*cfg)
		logrus.Infof("use config file from command line.")
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		logrus.Infof("search config file.")
	}
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	logrus.Infof("use config file:[%v]", viper.ConfigFileUsed())
	viper.WatchConfig()

	logrus.SetLevel(logrus.TraceLevel)
}
func main() {

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

}
