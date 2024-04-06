package service

import (
	"github.com/kentio/norn/internal/common"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	HTTPPort string `yaml:"http_port" json:"http_port"`

	// Webhook Secret(Optional)
	Github struct {
		Secret string `yaml:"secret" json:"secret"`
	}
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

// toMap converts Config struct to map
func (c *Config) toMap() map[string]interface{} {
	return common.StructToMap(c)
}

// Output to screen
func (c *Config) Output() {
	logrus.Info("Serve Config:")
	// each key-value pair in the map
	for k, v := range c.toMap() {
		logrus.Infof("%s: %v", k, v)
	}
}
