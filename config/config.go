package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

var ConfigPath string

type ConfigStruct struct {
	Patroni struct {
		Port string `yaml:"port"`
	} `yaml:"patroni"`
	Composition struct {
		Containers     []string `yaml:"container_list"`
		CompositionDir string   `yaml:"location"`
		Scripts        struct {
			StartCMD []string `yaml:"start"`
			StopCMD  []string `yaml:"stop"`
		} `yaml:"scripts"`
	} `yaml:"composition"`
	Cooldown float64 `yaml:"cooldown"`
	LogLevel string  `yaml:"log_level"`
}

type Config ConfigStruct

func New() (Config, error) {
	var conf Config
	file, err := os.ReadFile(ConfigPath)
	if err != nil {
		return conf, err
	}
	err = yaml.Unmarshal(file, &conf)
	if err != nil {
		return conf, err
	}

	return conf, nil
}
