package merle

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

// Thing configuration.  All things share this configuration.
type thingConfig struct {
	// The thing
	Thing struct {
		// Thing ID.  IDs are unique within an application to
		// differenciate one thing from another.
		Id          string `yaml:"Id"`
		// Thing model
		Model       string `yaml:"Model"`
		// Thing name
		Name        string `yaml:"Name"`
		// System user allowed to view thing with web browser.  HTTP
		// basic attentication prompts for user/passwd.  The user must
		// be a valid user on the system.  Passwd is set using standard
		// system tools.
		User        string `yaml:"User"`
		// Port for public HTTP server to view the thing, typically :80
		PortPublic  uint   `yaml:"PortPublic"`
		// Port for private HTTP server.  The private port is used to
		// connect to thing's mother using a websocket.
		PortPrivate uint   `yaml:"PortPrivate"`
	} `yaml:"Thing"`
	// Every thing has a mother.  A Mother is also a thing.
	Mother struct {
		// Mother host address
		Host        string `yaml:"Host"`
		// User on host associated with key
		User        string `yaml:"User"`
		// Key is a file from which the identity (private key) for
		// public key authentication is read.  See ssh -i option for
		// more information.
		Key         string `yaml:"Key"`
		// Port on host for mother's private HTTP server
		PortPrivate uint   `yaml:"PortPrivate"`
	} `yaml:"Mother"`
}

// Configurator is an interface to thing configuration.  A thing get's it's
// configuration through this interface.
type Configurator interface {
	// Parse cofiguration into supplied interface.
	Parse(interface{}) error
}

// A YAML configurator
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
	thingCfg, ok := cfg.(*thingConfig)
	if ok {
		thingCfg.Thing.Id = c.id
		thingCfg.Thing.Model = c.model
		thingCfg.Thing.Name = c.name
		log.Printf("Config parsed: %+v", cfg)
	}
	return nil
}
