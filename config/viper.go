package config

import (
	"testUtils/fakeDevice/clog"
	"testUtils/fakeDevice/utils"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func InitViper(cfgFile *string) error {
	clog.Logger().Infof("配置文件初始化")
	viper.AutomaticEnv()
	viper.SetConfigType("yaml")
	if cfgFile != nil && *cfgFile != "" {
		viper.SetConfigFile(*cfgFile)
		clog.Logger().Infof("使用指定的配置文件:%s", *cfgFile)
	} else {
		currentDir, err := utils.GetCurrentDir()
		if err != nil {
			return err
		}
		viper.AddConfigPath(currentDir)
		viper.AddConfigPath(".")
		clog.Logger().Infof("自动搜索配置文件。")
	}
	if err := viper.ReadInConfig(); err != nil {
		return errors.Cause(err)
	}
	clog.Logger().Infof("初始化配置文件成功，使用配置文件:%s", viper.ConfigFileUsed())
	viper.OnConfigChange(func(e fsnotify.Event) {
		logrus.Warnf("配置文件 %s 因为 %s 被变更,重新加载。", e.Name, e.Op)
	})
	viper.WatchConfig()
	return nil
}
