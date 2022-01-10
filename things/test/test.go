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

func (t *test) cb(p *merle.Packet) {
}

func (t *test) Subscribe() merle.Subscribers {
	return merle.Subscribers{
		{"msg", t.cb},
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

func (t *test) Run(p *merle.Packet) {
	t.log.Println("run")
	for {
	}
}
