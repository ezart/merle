// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

import (
	"fmt"
	"log"
	"math/rand"
	"os/exec"
	"syscall"
	"time"
)

func (t *Thing) _tunnelCreate() {

	var remote string
	var port string

	rand.Seed(time.Now().UnixNano())

	for {

		// TODO Need to use golang.org/x/crypto/ssh instead of
		// TODO os/exec'ing these ssh calls.  Also, look into
		// TODO using golang.org/x/crypto/ssh on hub-side of
		// TODO merle for bespoke ssh server.

		// ssh -i <key> <user>@<hub> curl -s localhost:8080/port?id=xxx

		log.Printf("%s Getting port...[ssh -i %s %s@%s curl -s localhost:8080/port?id=%s]...",
			t.logPrefix(), t.hubKey, t.hubUser, t.hubHost, t.Id)
		cmd := exec.Command("ssh", "-i", t.hubKey,
			t.hubUser+"@"+t.hubHost,
			"curl", "-s", "localhost:8080/port?id="+t.Id)

		// If the parent process (this app) dies, kill the ssh cmd also
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Pdeathsig: syscall.SIGTERM,
		}

		stdoutStderr, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("%s Get port failed: %s, err %v",
				t.logPrefix(), stdoutStderr, err)
			goto again
		}

		port = string(stdoutStderr)
		if port == "no ports available" {
			log.Println(t.logPrefix(), "No ports available; trying again\n")
			goto again
		}
		if port == "port busy" {
			log.Println(t.logPrefix(), "Port is busy; trying again\n")
			goto again
		}
		log.Println(t.logPrefix(), "Got port", port)

		// ssh -o ExitOnForwardFailure=yes -CNT -i <key> -R 8081:localhost:8080 <hub>
		//
		//  (The ExitOnForwardFailure=yes is to exit ssh if the remote port forwarding fails,
		//   most likely from port already being in-use on the server side).

		remote = fmt.Sprintf("%s:localhost:8080", port)
		log.Printf("%s Creating tunnel...[ssh -o ExitOnForwardFailure=yes -CNT -i %s -R %s %s@%s]...",
			t.logPrefix(), t.hubKey, remote, t.hubUser, t.hubHost)
		cmd = exec.Command("ssh", "-o", "ExitOnForwardFailure=yes",
			"-CNT", "-i", t.hubKey,
			"-R", remote, t.hubUser+"@"+t.hubHost)

		// If the parent process (this app) dies, kill the ssh cmd also
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Pdeathsig: syscall.SIGTERM,
		}

		stdoutStderr, err = cmd.CombinedOutput()
		if err != nil {
			log.Printf("%s Create tunnel failed: %s, err %v",
				t.logPrefix(), stdoutStderr, err)
		}

	again:
		// TODO maybe try some exponential back-off aglo ala TCP

		// Sleep for some number of random seconds between 1 and 10
		// before trying (again).  This will keep us from grinding
		// the CPU trying to connect all the time, and in the case
		// of multi clients starting at exactly the same time will
		// avoid port contention.

		f := rand.Float32() * 10
		log.Printf("%s Sleeping for %f seconds", t.logPrefix(), f)
		time.Sleep(time.Duration(f*1000) * time.Millisecond)
	}
}

func (t *Thing) tunnelCreate() {

	if t.hubHost == "" || t.hubUser == "" || t.hubKey == "" {
		log.Println(t.logPrefix(),
			"Skipping tunnel; missing hub connection parameters")
		return
	}

	go t._tunnelCreate()
}

// Configure tunnel
func (t *Thing) TunnelConfig(host, user, key string) {
	t.hubHost = host
	t.hubUser = user
	t.hubKey = key
}
