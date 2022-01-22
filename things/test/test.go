package test

import (
	"github.com/scottfeldman/merle"
	"log"
)

type thing struct {
	log *log.Logger
}

func NewModel(tlog *log.Logger, demo bool) merle.Thinger {
	return &thing{log: tlog}
}

func (t *thing) run(p *merle.Packet) {
	select{}
}

func (t *thing) Subscribe() merle.Subscribers {
	return merle.Subscribers{
		"CmdRun": t.run,
	}
}

type cfg struct {
	Test struct {
		Test string `yaml:"Test"`
	} `yaml:"Test"`
}

func (t *thing) Config(config merle.Configurator) error {
	var cfg cfg
	if err := config.Parse(&cfg); err != nil {
		return err
	}
	return nil
}

func (t *thing) Template() string {
	return "web/templates/test.html"
}
