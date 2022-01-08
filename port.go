package merle

import (
)

type ports struct {
	max uint
	match string
}

func NewPorts(max uint, match string) *ports {
	return &ports{
		max: max,
		match: match,
	}
}

func (p *ports) Start() {
	/*
	if err := p.initPorts(max); err != nil {
		return err
	}

	ticker := time.NewTicker(time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				if err := _portScan(match, cb); err != nil {
					return
				}
			}
		}
	}()

	return nil
	*/
}

func (p *ports) Stop() {
}
