// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

//go:build !tinygo
// +build !tinygo

package merle

import (
	"fmt"
	"math/rand"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

// Tunnel (remote SSH port forwarding) to connect a child thing to it's mother thing
type tunnel struct {
	thing       *Thing
	host        string
	user        string
	portPrivate uint
	portRemote  uint
}

func newTunnel(t *Thing, host, user string,
	portPrivate, portRemote uint) *tunnel {
	return &tunnel{
		thing:       t,
		host:        host,
		user:        user,
		portPrivate: portPrivate,
		portRemote:  portRemote,
	}
}

// TODO Need to use golang.org/x/crypto/ssh instead of
// TODO os/exec'ing these ssh calls.  Also, look into
// TODO using golang.org/x/crypto/ssh on hub-side of
// TODO merle for bespoke ssh server.

func (t *tunnel) getPort() string {

	// ssh <user>@<host> curl -s localhost:<privatePort>/port/<id>

	privatePort := strconv.FormatUint(uint64(t.portRemote), 10)

	args := []string{
		t.user + "@" + t.host,
		"curl", "-s",
		"localhost:" + privatePort + "/port/" + t.thing.id,
	}

	t.thing.log.printf("Tunnel getting port [ssh %s]", args)

	cmd := exec.Command("ssh", args...)

	// If the parent process (this app) dies, kill the ssh cmd also
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGTERM,
	}

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		t.thing.log.printf("Tunnel get port failed: %s, err %v", stdoutStderr, err)
		return ""
	}

	port := string(stdoutStderr)

	switch port {
	case "404 page not found\n":
		t.thing.log.println("Tunnel weirdness; Thing trying to be its own Mother?; trying again")
		return ""
	case "no ports available":
		t.thing.log.println("Tunnel no ports available; trying again")
		return ""
	case "port busy":
		t.thing.log.println("Tunnel port is busy; trying again")
		return ""
	}

	return port
}

func (t *tunnel) tunnel(port string) error {

	// ssh -o ExitOnForwardFailure=yes -CNT -R 8081:localhost:8080 <hub>
	//
	//  (The ExitOnForwardFailure=yes is to exit ssh if the remote port forwarding fails,
	//   most likely from port already being in-use on the server side).

	remote := fmt.Sprintf("%s:localhost:%d", port, t.portPrivate)

	args := []string{
		"-CNT",
		"-o", "ExitOnForwardFailure=yes",
		"-R", remote, t.user + "@" + t.host,
	}

	t.thing.log.printf("Creating tunnel [ssh %s]", args)

	cmd := exec.Command("ssh", args...)

	// If the parent process (this app) dies, kill the ssh cmd also
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGTERM,
	}

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		t.thing.log.printf("Create tunnel failed: %s, err %v", stdoutStderr, err)
	}

	return err
}

func (t *tunnel) create() {
	var err error
	var port string

	rand.Seed(time.Now().UnixNano())

	for {

		port = t.getPort()
		if port == "" {
			goto again
		}

		t.thing.log.println("Tunnel got port", port)

		err = t.tunnel(port)
		if err != nil {
			goto again
		}

		t.thing.log.println("Tunnel disconnected")

	again:
		// TODO maybe try some exponential back-off aglo ala TCP

		// Sleep for some number of random seconds between 1 and 10
		// before trying (again).  This will keep us from grinding
		// the CPU trying to connect all the time, and in the case
		// of multi clients starting at exactly the same time will
		// avoid port contention.

		f := rand.Float32() * 10
		t.thing.log.printf("Tunnel create sleeping for %f seconds", f)
		time.Sleep(time.Duration(f*1000) * time.Millisecond)
	}
}

func (t *tunnel) start() {
	if t.host == "" {
		t.thing.log.println("Skipping tunnel to mother; missing host")
		return
	}

	if t.user == "" {
		t.thing.log.println("Skipping tunnel to mother; missing user")
		return
	}

	if t.portRemote == 0 {
		t.thing.log.println("Skipping tunnel to mother; missing remote port")
		return
	}

	if t.portPrivate == 0 {
		t.thing.log.println("Skipping tunnel to mother; missing private port")
		return
	}

	go t.create()
}

func (t *tunnel) stop() {
}
