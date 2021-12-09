package merle

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

const testId = "HS30-01132" // sorry if that's your car
const testModel = "240z"
const testName = "Fairlady"
const helloWorld = "Hello World!"

func newDevice(m IModel) *Device {
	var inHub = false
	var status = "online"
	var startupTime = time.Now()

	return NewDevice(m, inHub, testId, testModel,
		testName, status, startupTime)
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

	err := d.Run("", 80, 8080, "", "", "")
	if err.Error() != "Device Run() exited unexpectedly" {
		t.Errorf("Run failed: %s", err)
	}
}

type simple struct {
	done chan (bool)
}

func (s *simple) Init(d *Device) error {
	s.done = make(chan (bool))
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
	fmt.Fprintf(w, helloWorld)
}

func checkIdentifyResp(r *MsgIdentifyResp) error {
	if r.Id != testId ||
		r.Model != testModel ||
		r.Name != testName {
		return fmt.Errorf("Identify not matching")
	}
	return nil
}

func testIdentify(t *testing.T, httpPort int) {
	var p = &port{port: httpPort}

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

	p.ws.Close()
}

func testDone(t *testing.T, httpPort int) {
	var p = &port{port: httpPort}

	err := p.wsOpen()
	if err != nil {
		t.Errorf("ws open failed: %s", err)
	}

	// Send msg to shutdown device
	var msg = MsgCmd{Type: "cmd", Cmd: "Done"}

	err = p.writeJSON(msg)
	if err != nil {
		t.Errorf("ws writeJSON failed: %s", err)
	}

	p.ws.Close()
}

func testHomePage(t *testing.T, httpPort int) {
	addr := ":" + strconv.Itoa(httpPort)

	get, err := http.Get("http://localhost" + addr)
	if err != nil {
		t.Errorf("Get http://localhost%s failed: %s", addr, err)
	}

	body, err := io.ReadAll(get.Body)
	get.Body.Close()

	if err != nil {
		t.Errorf("Get localhost%s failed: %s", addr, err)
	}

	contents := strings.TrimSpace(string(body))
	if contents != helloWorld {
		t.Errorf("Get localhost%s body failed.  Got: %s, wanted %s",
			addr, contents, helloWorld)
	}
}

func testSimple(t *testing.T, publicPort, privatePort int) {
	// sleep a second for http servers to start
	time.Sleep(time.Second)
	testHomePage(t, publicPort)
	testIdentify(t, publicPort)
	testIdentify(t, privatePort)
	testDone(t, privatePort)
}

func TestSimple(t *testing.T) {
	var s simple
	var portPublic = 8081
	var portPrivate = 8080

	d := newDevice(&s)
	if d == nil {
		t.Errorf("Create new device failed")
	}

	go testSimple(t, portPublic, portPrivate)

	err := d.Run("testtest", portPublic, portPrivate, "", "", "")
	if err.Error() != "Device Run() exited unexpectedly" {
		t.Errorf("Run failed: %s", err)
	}
}
