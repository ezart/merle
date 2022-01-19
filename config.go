package merle

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

// Thing configuration.  All Things share this configuration.
type ThingConfig struct {

	// The section describes a Thing.
	Thing struct {
		// (Optional) Thing's Id.  Ids are unique within an application
		// to differenciate one Thing from another.  Id is optional; if
		// Id is not given, a system-wide unique Id is assigned.
		Id string `yaml:"Id"`
		// Thing's Model.  Should match one of the models support by
		// Merle.  See merle --models for list of support models.
		Model string `yaml:"Model"`
		// Thing's Name
		Name string `yaml:"Name"`
		// (Optional) system User.  If a User is given, any browser
		// views of the Thing's home page  will prompt for user/passwd.
		// HTTP Basic Authentication is used and the user/passwd given
		// must match the system creditials for the User.  If no User
		// is given, HTTP Basic Authentication is skipped; anyone can
		// view the home page.
		User string `yaml:"User"`
		// (Optional) If PortPublic is given, an HTTP web server is
		// started on port PortPublic.  PortPublic is typically set to
		// 80.  The HTTP web server runs the Thing's home page.
		PortPublic uint `yaml:"PortPublic"`
		// (Optional) If PortPublicTLS is given, an HTTPS web server is
		// started on port PortPublicTLS.  PortPublicTLS is typically
		// set to 443.  The HTTPS web server will self-certify using a
		// certificate from Let's Encrypt.  The public HTTPS server
		// will securely run the Thing's home page.  If PortPublicTLS
		// is given, PortPublic must be given.
		PortPublicTLS uint `yaml:"PortPublicTLS"`
		// (Optional) If PortPrivate is given, a HTTP server is
		// started on port PortPrivate.  This HTTP server does not
		// server up the Thing's home page but rather connects to
		// Thing's Mother using a websocket over HTTP.
		PortPrivate uint `yaml:"PortPrivate"`
	} `yaml:"Thing"`

	// (Optional) This section describes a Thing's Mother.  Every Thing has
	// a Mother.  A Mother is also a Thing, so we can build a hierarchy of
	// Things, with a Thing having potentially a Mother, a GrandMother, a
	// Great GrandMother, etc.
	Mother struct {
		// Mother's Host address.  Host, User and Key are used to
		// connect this Thing to it's Mother using a SSH connection.
		// For example: ssh -i <Key> <User>@<Host>.
		Host string `yaml:"Host"`
		// User on Host associated with Key
		User string `yaml:"User"`
		// Key is the file path of the SSH identity key.  See ssh -i
		// option for more information.
		Key string `yaml:"Key"`
		// Port on Host for Mother's private HTTP server
		PortPrivate uint `yaml:"PortPrivate"`
	} `yaml:"Mother"`
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
