package test

import (
	"github.com/scottfeldman/merle"
)

type test struct {
}

func NewTest() merle.IThing {
	return &test{}
}

func (t *test) cb() {
}

func (t *test) Subscribe() merle.Subscribers {
	return merle.Subscribers{
		"msg": t.cb,
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
