package merle

import (
	"fmt"
	"log"
	"gopkg.in/yaml.v3"
	"os"
)

type ThingConfig struct {
	Thing struct {
		Id          string `yaml:"Id"`
		Model       string `yaml:"Model"`
		Name        string `yaml:"Name"`
		User        string `yaml:"User"`
		PortPublic  uint   `yaml:"PortPublic"`
		PortPrivate uint   `yaml:"PortPrivate"`
	} `yaml:"Thing"`
	Mother struct {
		Host        string `yaml:"Host"`
		User        string `yaml:"User"`
		Key         string `yaml:"Key"`
		PortPrivate uint   `yaml:"PortPrivate"`
	} `yaml:"Mother"`
}

type Configurator interface {
	Parse(interface{}) error
}

type yamlConfig struct {
	file string
}

func NewYamlConfig(file string) Configurator {
	return &yamlConfig{file: file}
}

func (c *yamlConfig) Parse(cfg interface{}) error {
	f, err := os.Open(c.file)
	if err != nil {
		return fmt.Errorf("Opening config file failure: %s", err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		return fmt.Errorf("Config decode error: %s", err)
	}

	log.Printf("%+v", cfg)
	return nil
}
