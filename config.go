package merle

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"flag"
)

func FlagThingConfig(id, model, name, user, assetsDir string) *ThingConfig {
	var cfg ThingConfig

	flag.BoolVar(&cfg.Thing.Prime, "prime", false, "Run as Thing-prime")
	flag.UintVar(&cfg.Thing.PortPrime, "pport", 0, "Prime Port")

	flag.StringVar(&cfg.Thing.Id, "id", id, "Thing ID")
	flag.StringVar(&cfg.Thing.Model, "model", model, "Thing model")
	flag.StringVar(&cfg.Thing.Name, "name", name, "Thing name")

	flag.StringVar(&cfg.Thing.User, "luser", user,
		"Local user for HTTP Basic Authentication")
	flag.UintVar(&cfg.Thing.PortPublic, "lport", 80,
		"Local public HTTP listening port")
	flag.UintVar(&cfg.Thing.PortPublicTLS, "lportTLS", 0,
		"Local public HTTPS listening port (default 0, but usually 443)")
	flag.UintVar(&cfg.Thing.PortPrivate, "lportPriv", 8080,
		"Local private HTTP listening port")
	flag.StringVar(&cfg.Thing.AssetsDir, "lassets", assetsDir,
		"Local path to assets directory")

	flag.StringVar(&cfg.Mother.Host, "rhost", "",
		"Remote host name or IP address")
	flag.StringVar(&cfg.Mother.User, "ruser", user,
		"Remote user")
	flag.StringVar(&cfg.Mother.Key, "rkey",
		"/home/" + user + "/.ssh/id_rsa", "Remote SSH identity key")
	flag.UintVar(&cfg.Mother.PortPrivate, "rport", 8080,
		"Remote private HTTP listening port")

	return &cfg
}

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
