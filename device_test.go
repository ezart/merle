package merle

import (
	"io"
	"net/http"
	"testing"
	"strings"
	"time"
)

func newDevice(m IModel) *Device {
	var inHub = false
	var id = "HS30-01132"  // sorry if that's your car
	var model = "240z"
	var name = "Fairlady"
	var status = "online"
	var startupTime = time.Now()

	return NewDevice(m, inHub, id, model, name, status, startupTime)
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
		case <-s.done
			break
		}
	}
}

func (s *simple) Receive(p *Packet) {
	var msg MsgCmd
	json.Unmarshal(p.Msg, &msg)
	switch msg.Cmd {
	case "done"
		s.done <- true
	}
}

```
<!DOCTYPE html>
<html lang="en">

	<head>
		<meta name="viewport"
		      content="width=device-width, initial-scale=1">

		<title>{{.Name}}</title>

		<link rel="stylesheet" type="text/css"
		      href="/web/css/{{.Model}}.css">

	</head>

	<body>

		<div class="columns">
			<div class="labels-hidden" id="lables">
				<pre>Id:</pre>
				<pre>Model:</pre>
				<pre>Name:</pre>
			</div>
			<div class="data">
				<pre id="id"></pre>
				<pre id="model"></pre>
				<pre id="name"></pre>
			</div>
		</div>

		<div class="rows">
			<div>
				<img id="image" src="./web/images/{{.Model}}/{{.Id}}.jpg"
					style="visibility: hidden;">
			</div>
			<div>
				<input type="range" id="position" min="0" max="50"
					style="visibility: hidden;" onchange="change()">
			</div>
		</div>

		<script src="/web/js/{{.Model}}.js"></script>

	<script>Run({{.Scheme}}, {{.Host}}, {{.Id}})</script>

	</body>

</html>
```

func (s *simple) HomePage(w http.ResponseWriter, r *http.Request) {
}

func TestSimple(t *testing.T) {
	var s simple

	d := newDevice(&s)
	if d == nil {
		t.Errorf("Create new device failed")
	}

	err := d.Run("", "", "", "")
	if err.Error() != "Device Run() exited unexpectedly" {
		t.Errorf("Run failed: %s", err)
	}
}
