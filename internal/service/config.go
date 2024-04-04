package service

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	HTTPPort string `yaml:"http_port" json:"http_port"`
}

func NewConfig() (*Config, error) {
	conf := &Config{}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("norn")
	viper.SetConfigType("yaml")
	viper.SetConfigFile("config.yaml")

	err := viper.ReadInConfig()
	if err != nil {
		logrus.Errorf("read config error: %v", err)
	}

	err = viper.Unmarshal(conf)
	if err != nil {
		logrus.Errorf("unable to decode into struct, %v", err)
		return nil, err
	}

	return conf, nil

}
