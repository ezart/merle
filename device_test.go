package merle

import (
	"net/http"
	"testing"
	"time"
)

type model struct {
}

func (m *model) Init(d *Device) error {
	return nil
}

func (m *model) Run() {
}

func (m *model) Receive(p *Packet) {
}

func HomePage(w http.ResponseWrite, r *http.Request) {
}

func TestSomething(t *testing.T) {
	var m model
	var inHub = false
	var id = "HS30-01132"  // sorry if that's your car
	var model = "240z"
	var name = "Fairlady"
	var status = "online"
	var startupTime = time.Now()

	d = NewDevice(&m, inHub, id, model, name, status, startupTime)
	if d == nil {
		t.Errorf("Create new device failed")
	}

	err := d.Run("", "", "", "")
	if err != nil {
		t.Errorf("Run failed: %s", err)
	}
}
