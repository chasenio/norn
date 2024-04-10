package service

import (
	"crypto/rsa"
	"fmt"
	"github.com/kentio/norn/internal/common"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"reflect"
	"strings"
)

type GithubConfig struct {
	// Secret is the webhook secret, it is used to verify the payload
	Secret string `yaml:"secret" json:"secret"`
	// AppID is the GitHub App ID
	AppID string `yaml:"app_id" json:"app_id"`
	// InstallationID is the GitHub App Installation ID
	// Usually, it from the webhook payload installation field { id: 123456 }
	//InstallationID string `yaml:"installation_id" json:"installation_id"`
	// PrivateKey is the GitHub App Private Key
	PrivateKey string `yaml:"private_key" json:"private_key"`
}

const (
	ConfigType  = "yaml"
	ConfigFile  = "config.yaml"
	DefaultPort = "8080"
	Debug       = true
)

type Config struct {
	HTTPPort string `yaml:"http_port" json:"http_port"`

	// Github is the GitHub configuration
	Github GithubConfig

	// Pick Path of Branchs
	Branches []string `yaml:"branches" json:"branches"`
	// Dev is Debug Mode
	Dev bool `yaml:"dev" json:"dev"`
}

func NewConfig() (*Config, error) {

	//githubCfg := GithubConfig{}
	viper.SetEnvPrefix("norn")
	viper.AutomaticEnv()
	viper.AllowEmptyEnv(true)
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetConfigType(ConfigType)
	viper.SetConfigFile(ConfigFile)
	//viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	if err := viper.ReadInConfig(); err == nil {
		logrus.Infof("found config file...")
		viper.ConfigFileUsed()
	}

	cfg := Config{
		HTTPPort: DefaultPort,
		Dev:      Debug,
	}

	BindStructEnv(cfg, "") // Bind Config in env

	if err := viper.Unmarshal(&cfg); err != nil {
		logrus.Errorf("unable to decode into struct, %v", err)
		return nil, err
	}

	cfg.Output()
	return &cfg, nil

}

// BindStructEnv Bind Config in env
func BindStructEnv(config interface{}, prefix string) {
	r := reflect.TypeOf(config)

	// iterate over each field for the type
	for j := 0; j < r.NumField(); j++ {
		field := r.Field(j)
		var name string
		if len(prefix) > 0 {
			name = fmt.Sprintf("%s.%s", prefix, field.Name)
		} else {
			name = field.Name
		}
		// if type is struct
		if field.Type.Kind() == reflect.Struct {
			// get field value
			value := reflect.ValueOf(config).FieldByName(field.Name).Interface()
			BindStructEnv(value, name)
			continue
		}

		// bind environment variables
		if err := viper.BindEnv(name); err != nil {
			logrus.Errorf("Bind Err：%v", err)
		}

	}
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
		logrus.Infof("\t%s: \t%v", k, v)
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
