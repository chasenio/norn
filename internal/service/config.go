package service

import (
	"crypto/rsa"
	"github.com/kentio/norn/internal/common"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type GithubConfig struct {
	// Secret is the webhook secret, it is used to verify the payload
	Secret string `yaml:"secret" json:"secret"`
	// AppID is the GitHub App ID
	AppID string `yaml:"app_id" json:"app_id"`
	// InstallationID is the GitHub App Installation ID
	// Usually, it from the webhook payload installation field { id: 123456 }
	InstallationID string `yaml:"installation_id" json:"installation_id"`
	// PrivateKey is the GitHub App Private Key
	PrivateKey string `yaml:"private_key" json:"private_key"`
}

type Config struct {
	HTTPPort string `yaml:"http_port" json:"http_port"`

	// Github is the GitHub configuration
	Github *GithubConfig

	// Pick Path of Branchs
	Branches []string `yaml:"branches" json:"branches"`
	// Dev is Debug Mode
	Dev bool
}

func NewConfig() (*Config, error) {
	cfg := &Config{}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("norn")
	viper.SetConfigType("yaml")
	viper.SetConfigFile("config.yaml")

	err := viper.ReadInConfig()
	if err != nil {
		logrus.Errorf("read config error: %v", err)
	}

	err = viper.Unmarshal(cfg)
	if err != nil {
		logrus.Errorf("unable to decode into struct, %v", err)
		return nil, err
	}

	cfg.Output()
	return cfg, nil

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

// PrivateKey returns the private key for the GitHub App
func (c *Config) PrivateKey() (*rsa.PrivateKey, error) {
	key, err := common.ToPrivateKeys(c.Github.PrivateKey)
	if err != nil {
		logrus.Fatalf("could not parse private key: %s", err)
		return nil, err
	}
	return key, nil
}
