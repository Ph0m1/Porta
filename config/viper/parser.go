// defines a config paser implementation based on the viper pkg
package viper

import (
	"fmt"

	"github.com/ph0m1/porta/config"
	"github.com/spf13/viper"
)

func New() config.Parser {
	return parser{viper.New()}
}

type parser struct {
	viper *viper.Viper
}

func (p parser) Parse(configFile string) (config.ServiceConfig, error) {
	p.viper.SetConfigFile(configFile)
	p.viper.AutomaticEnv()
	var cfg config.ServiceConfig
	if err := p.viper.ReadInConfig(); err != nil {
		return cfg, fmt.Errorf("Fatal error config file: %s\n", err)
	}
	if err := p.viper.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("Fatal error unmarshalling config file: %s\n", err)
	}
	if err := cfg.Init(); err != nil {
		return cfg, err
	}
	return cfg, nil
}
