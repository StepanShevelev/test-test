package config

import (
	"fmt"
	"github.com/spf13/viper"
	"time"
)

type Config struct {
	Host         string        `mapstructure:"host"`
	Port         string        `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"readTimeout"`
	WriteTimeout time.Duration `mapstructure:"writeTimeout"`
	Limit        int           `mapstructure:"limit"`
	Req          int           `mapstructure:"req"`
}

func New() *Config {
	return &Config{}
}

func (c *Config) Load(path string, name string, _type string) error {
	viper.AddConfigPath(path)
	viper.SetConfigName(name)
	viper.SetConfigType(_type)

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("read config error: %w", err)
	}

	err = viper.Unmarshal(c)

	if err != nil {
		return fmt.Errorf("unmarshalling config error: %w", err)
	}
	return nil
}
