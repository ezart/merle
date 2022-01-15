package test

import (
	"github.com/scottfeldman/merle"
	"log"
)

type test struct {
	log *log.Logger
}

func NewModel(tlog *log.Logger, demo bool) merle.Thinger {
	return &test{log: tlog}
}

func (t *test) run(p *merle.Packet) {
	select{}
}

func (t *test) Subscribe() merle.Subscribers {
	return merle.Subscribers{
		{"CmdRun", t.run},
	}
}

type cfg struct {
	Test struct {
		Test string `yaml:"Test"`
	} `yaml:"Test"`
}

func (t *test) Config(config merle.Configurator) error {
	var cfg cfg
	if err := config.Parse(&cfg); err != nil {
		return err
	}
	return nil
}

func (t *test) Template() string {
	return "web/templates/test.html"
}
