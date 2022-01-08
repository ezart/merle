// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

import (
	"log"
	"fmt"
	"math/rand"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

type tunnel struct {
	id string
	host string
	user string
	key string
	portRemote uint
}

func NewTunnel(id, host, user, key string, portRemote uint) *tunnel {
	return &tunnel{
		id: id,
		host: host,
		user: user,
		key: key,
		portRemote: portRemote,
	}
}

// TODO Need to use golang.org/x/crypto/ssh instead of
// TODO os/exec'ing these ssh calls.  Also, look into
// TODO using golang.org/x/crypto/ssh on hub-side of
// TODO merle for bespoke ssh server.

func (t *tunnel) getPort() string {
	args := []string{}

	// ssh -i <key> <user>@<host> curl -s localhost:<privatePort>/port/<id>

	args = append(args, "-i", t.key)
	args = append(args, t.user+"@"+t.host)
	args = append(args, "curl", "-s")
	args = append(args, "localhost:"+
		strconv.FormatUint(uint64(t.portRemote), 10)+
		"/port/"+t.id)

	log.Printf("Tunnel getting port [ssh %s]...", args)

	cmd := exec.Command("ssh", args...)

	// If the parent process (this app) dies, kill the ssh cmd also
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGTERM,
	}

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Tunnel get port failed: %s, err %v", stdoutStderr, err)
		return ""
	}

	port := string(stdoutStderr)

	switch port {
	case "no ports available":
		log.Println("Tunnel no ports available; trying again\n")
		return ""
	case "port busy":
		log.Println("Tunnel port is busy; trying again\n")
		return ""
	}

	return port
}

func (t *tunnel) tunnel(port string) error {

	args := []string{}

	// ssh -o ExitOnForwardFailure=yes -CNT -i <key> -R 8081:localhost:8080 <hub>
	//
	//  (The ExitOnForwardFailure=yes is to exit ssh if the remote port forwarding fails,
	//   most likely from port already being in-use on the server side).

	remote := fmt.Sprintf("%s:localhost:%d", port, t.portRemote)

	args = append(args, "-CNT")
	args = append(args, "-i", t.key)
	args = append(args, "-o", "ExitOnForwardFailure=yes")
	args = append(args, "-R", remote, t.user+"@"+t.host)

	log.Printf("Creating tunnel [ssh %s]...", args)

	cmd := exec.Command("ssh", args...)

	// If the parent process (this app) dies, kill the ssh cmd also
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGTERM,
	}

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Create tunnel failed: %s, err %v", stdoutStderr, err)
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

		log.Println("Tunnel create got port", port)

		err = t.tunnel(port)
		if err != nil {
			goto again
		}

		log.Println("Tunnel disconnected")

	again:
		// TODO maybe try some exponential back-off aglo ala TCP

		// Sleep for some number of random seconds between 1 and 10
		// before trying (again).  This will keep us from grinding
		// the CPU trying to connect all the time, and in the case
		// of multi clients starting at exactly the same time will
		// avoid port contention.

		f := rand.Float32() * 10
		log.Printf("Tunnel create sleeping for %f seconds", f)
		time.Sleep(time.Duration(f*1000) * time.Millisecond)
	}
}

func (t *tunnel) Start() {
	if t.host == "" {
		log.Println("Skipping tunnel; missing host")
		return
	}

	if t.user == "" {
		log.Println("Skipping tunnel; missing user")
		return
	}

	if t.key == "" {
		log.Println("Skipping tunnel; missing key")
		return
	}

	if t.portRemote == 0 {
		log.Println("Skipping tunnel; missing remote port")
		return
	}

	go t.create()
}

func (t *tunnel) Stop() {
}
