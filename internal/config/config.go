package config

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
)

type Service struct {
	Address string `yaml:"address"`
}

type Config struct {
	Bot struct {
		Token   string `yaml:"token"`
		Name    string `yaml:"name"`
		Address struct {
			Host string `yaml:"host"`
			Port string `yaml:"port"`
		}
	} `yaml:"bot"`
	Database struct {
		Tarantool struct {
			Address  string `yaml:"address"`
			User     string `yaml:"user"`
			Password string `yaml:"password"`
		} `yaml:"tarantool"`
	} `yaml:"database"`

	Mattermost struct {
		Url      string `yaml:"url"`
		TeamName string `yaml:"teamName"`
	} `yaml:"mattermost"`
}

func LoadConfig(path, mode string) *Config {
	viper.SetConfigName("config." + mode)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("bad path to config: '%s', mode: %s", path, mode)
	}
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("ERROR: YAML parsing")
	}
	return &config
}

func (cnf *Config) GetHttpUrlBot() string {
	return fmt.Sprintf("http://%s:%s", cnf.Bot.Address.Host, cnf.Bot.Address.Port)
}
func (cnf *Config) GetUrlBot() string {
	return fmt.Sprintf("%s:%s", cnf.Bot.Address.Host, cnf.Bot.Address.Port)
}
