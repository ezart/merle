// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

import (
	"fmt"
	"math/rand"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

// TODO Need to use golang.org/x/crypto/ssh instead of
// TODO os/exec'ing these ssh calls.  Also, look into
// TODO using golang.org/x/crypto/ssh on hub-side of
// TODO merle for bespoke ssh server.

func (t *Thing) getPortFromMother() string {
	args := []string{}

	// ssh -i <key> <user>@<host> curl -s localhost:<privatePort>/port/<id>

	args = append(args, "-i", t.motherKey)
	args = append(args, t.motherUser + "@" + t.motherHost)
	args = append(args, "curl", "-s")
	args = append(args, "localhost:" + strconv.Itoa(t.motherPortPrivate) + "/port/" + t.id)

	t.log.Printf("Tunnel getting port [ssh %s]...", args)

	cmd := exec.Command("ssh", args...)

	// If the parent process (this app) dies, kill the ssh cmd also
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGTERM,
	}

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		t.log.Printf("Tunnel get port failed: %s, err %v", stdoutStderr, err)
		return ""
	}

	port := string(stdoutStderr)

	switch port {
	case "no ports available":
		t.log.Println("Tunnel no ports available; trying again\n")
		return ""
	case "port busy":
		t.log.Println("Tunnel port is busy; trying again\n")
		return ""
	}

	return port
}

func (t *Thing) tunnelToMother(port string) error {

	args := []string{}

	// ssh -o ExitOnForwardFailure=yes -CNT -i <key> -R 8081:localhost:8080 <hub>
	//
	//  (The ExitOnForwardFailure=yes is to exit ssh if the remote port forwarding fails,
	//   most likely from port already being in-use on the server side).

	remote := fmt.Sprintf("%s:localhost:%d", port, t.portPrivate)

	args = append(args, "-CNT")
	args = append(args, "-i", t.motherKey)
	args = append(args, "-o", "ExitOnForwardFailure=yes")
	args = append(args, "-R", remote, t.motherUser+"@"+t.motherHost)

	t.log.Printf("Creating tunnel [ssh %s]...", args)

	cmd := exec.Command("ssh", args...)

	// If the parent process (this app) dies, kill the ssh cmd also
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGTERM,
	}

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		t.log.Printf("Create tunnel failed: %s, err %v", stdoutStderr, err)
	}

	return err
}

func (t *Thing) _tunnelCreate() {
	var err error
	var port string

	rand.Seed(time.Now().UnixNano())

	for {

		port = t.getPortFromMother()
		if port == "" {
			goto again
		}

		t.log.Println("Tunnel create got port", port)

		err = t.tunnelToMother(port)
		if err != nil {
			goto again
		}

		t.log.Println("Tunnel disconnected")

	again:
		// TODO maybe try some exponential back-off aglo ala TCP

		// Sleep for some number of random seconds between 1 and 10
		// before trying (again).  This will keep us from grinding
		// the CPU trying to connect all the time, and in the case
		// of multi clients starting at exactly the same time will
		// avoid port contention.

		f := rand.Float32() * 10
		t.log.Printf("Tunnel create sleeping for %f seconds", f)
		time.Sleep(time.Duration(f*1000) * time.Millisecond)
	}
}

func (t *Thing) tunnelCreate() {

	if t.motherHost == "" {
		t.log.Println("Skipping tunnel; missing Mother host")
		return
	}

	if t.motherUser == "" {
		t.log.Println("Skipping tunnel; missing Mother user")
		return
	}

	if t.motherKey == "" {
		t.log.Println("Skipping tunnel; missing Mother key")
		return
	}

	if t.motherPortPrivate == 0 {
		t.log.Println("Skipping tunnel; missing Mother private port")
		return
	}

	go t._tunnelCreate()
}

// Configure tunnel
func (t *Thing) TunnelConfig(host, user, key string, portPrivate int) {
	t.motherHost = host
	t.motherUser = user
	t.motherKey = key
	t.motherPortPrivate = portPrivate
}
