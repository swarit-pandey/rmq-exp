package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Conf struct {
	Core           *RootConf
	configFilePath string
	v              *viper.Viper
}

func New(configFilePath string) *Conf {
	return &Conf{
		Core:           &RootConf{},
		configFilePath: configFilePath,
		v:              viper.New(),
	}
}

func (c *Conf) Load() error {
	c.v.SetConfigType("yaml")
	c.v.SetConfigFile(c.configFilePath)

	if err := c.v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	return nil
}

func (c *Conf) Unmarshal() error {
	if err := c.v.Unmarshal(c); err != nil {
		return fmt.Errorf("failed to unmarhshal: %w", err)
	}

	return nil
}
