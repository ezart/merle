package merle

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

// Configurator is an interface to thing configuration.  A thing get's it's
// configuration through this interface.
type Configurator interface {
	// Parse cofiguration into supplied interface.
	Parse(interface{}) error
}

type yamlConfig struct {
	file string
}

// A YAML configurator
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

	log.Printf("Config parsed: %+v", cfg)
	return nil
}

// A child configurator used by the bridge when attaching a thing.
type childConfig struct {
	id    string
	model string
	name  string
}

func newChildConfig(id, model, name string) Configurator {
	return &childConfig{
		id:    id,
		model: model,
		name:  name,
	}
}

func (c *childConfig) Parse(cfg interface{}) error {
	thingCfg, ok := cfg.(*ThingConfig)
	if ok {
		thingCfg.Thing.Id = c.id
		thingCfg.Thing.Model = c.model
		thingCfg.Thing.Name = c.name
		log.Printf("Config parsed: %+v", cfg)
	}
	return nil
}
