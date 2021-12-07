package merle

import (
	"fmt"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"strings"
	"time"
)

const testId = "HS30-01132"  // sorry if that's your car
const testModel = "240z"
const testName = "Fairlady"

func newDevice(m IModel) *Device {
	var inHub = false
	var status = "online"
	var startupTime = time.Now()

	return NewDevice(m, inHub, testId, testModel,
		testName, status, startupTime)
}

func checkIdentifyResp(r *MsgIdentifyResp) error {
	if r.Id != testId ||
	   r.Model != testModel ||
	   r.Name != testName {
		return fmt.Errorf("Identify not matching")
	}
	return nil
}

type minimal struct {
}

func (m *minimal) Init(d *Device) error {
	return nil
}

func (m *minimal) Run() {
}

func (m *minimal) Receive(p *Packet) {
}

func (m *minimal) HomePage(w http.ResponseWriter, r *http.Request) {
}

func TestMinimalParams(t *testing.T) {
	var m minimal

	d := NewDevice(&m, false, "", "foo", "bar", "online", time.Now())
	if d == nil {
		t.Errorf("Create new device failed")
	}

	if d.Id() == "" {
		t.Errorf("d.Id() empty string")
	}

	d = NewDevice(&m, false, "", "", "bar", "online", time.Now())
	if d != nil {
		t.Errorf("Create new device succeeded with model=''")
	}

	d = NewDevice(&m, false, "", "foo", "", "online", time.Now())
	if d != nil {
		t.Errorf("Create new device succeeded with name=''")
	}

	d = NewDevice(&m, false, "", "foo", "bar", "", time.Now())
	if d != nil {
		t.Errorf("Create new device succeeded with status=''")
	}
}

func TestMinimalRun(t *testing.T) {
	var m minimal

	d := newDevice(&m)
	if d == nil {
		t.Errorf("Create new device failed")
	}

	err := d.Run("", "", "", "")
	if err.Error() != "Device Run() exited unexpectedly" {
		t.Errorf("Run failed: %s", err)
	}

	// sleep a second for http server to start
	time.Sleep(time.Second)

	// Since authUser="" was passed into d.Run(), only expecting
	// http server to be running on localhost:8080.  localhost:80
	// should not be listening.  404 should be returned for
	// localhost:8080 since no HomePage.

	resp, err := http.Get("http://localhost:8080")
	if err != nil {
		t.Errorf("Get failed: %s", err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Errorf("Get localhost:8080 failed: %s", err)
	}
	if strings.TrimSpace(string(body)) != "404 page not found" {
		t.Errorf("Get localhost:8080 expected '404 page not found', got: %s foo", body)
	}

	resp, err = http.Get("http://localhost:80")
	if err == nil {
		resp.Body.Close()
		t.Errorf("Get localhost:80 didn't fail")
	}
}

type simple struct {
	done chan(bool)
}

func (s *simple) Init(d *Device) error {
	s.done = make(chan(bool))
	return nil
}

func (s *simple) Run() {
	for {
		select {
		case <-s.done:
			return
		}
	}
}

func (s *simple) Receive(p *Packet) {
	var msg MsgCmd
	json.Unmarshal(p.Msg, &msg)
	switch msg.Cmd {
	case "Done":
		s.done <- true
	}
}

func (s *simple) HomePage(w http.ResponseWriter, r *http.Request) {
}

func runWs(t *testing.T) {
	var p = &port{port: 8080}

	// sleep a second for http server to start
	time.Sleep(time.Second)

	err := p.wsOpen()
	if err != nil {
		t.Errorf("ws open failed: %s", err)
	}

	err = p.wsIdentify()
	if err != nil {
		t.Errorf("ws identify failed: %s", err)
	}

	resp, err := p.wsIdentifyResp()
	if err != nil {
		t.Errorf("ws identify response failed: %s", err)
	}

	err = checkIdentifyResp(resp)
	if err != nil {
		t.Errorf("Unexpected identify response: %s", err)
	}

	var msg = MsgCmd{Type: "cmd", Cmd: "Done"}

	err = p.writeJSON(msg)
	if err != nil {
		t.Errorf("ws writeJSON failed: %s", err)
	}
}

func TestSimple(t *testing.T) {
	var s simple

	d := newDevice(&s)
	if d == nil {
		t.Errorf("Create new device failed")
	}

	go runWs(t)

	err := d.Run("", "", "", "")
	if err.Error() != "Device Run() exited unexpectedly" {
		t.Errorf("Run failed: %s", err)
	}
}
