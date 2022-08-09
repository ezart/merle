// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

//go:build !tinygo
// +build !tinygo

package merle

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// Tunnel (remote SSH port forwarding) to connect a child Thing to it's mother Thing
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

func getRemote(user, server string) (*ssh.Client, error) {
	hostKeyCallback, err := knownhosts.New("/home/" + user + "/.ssh/known_hosts")
	if err != nil {
		return nil, err
	}

	// TODO: Allow different key name to be passed in thing.Cfg?
	// TODO: Currently hardcoded to id_ras

	key, err := ioutil.ReadFile("/home/" + user + "/.ssh/id_rsa")
	if err != nil {
		return nil, fmt.Errorf("Unable to read private key: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: hostKeyCallback,
	}

	client, err := ssh.Dial("tcp", server+":22", config)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect: %v", err)
	}

	return client, nil
}

func (t *tunnel) getPort() (string, error) {

	// ssh <user>@<host> curl -s localhost:<privatePort>/port/<id>

	privatePort := strconv.FormatUint(uint64(t.portRemote), 10)
	cmd := "curl -s localhost:" + privatePort + "/port/" + t.thing.id

	t.thing.log.printf("Tunnel getting port [ssh %s@%s %s]",
		t.user, t.host, cmd)

	remote, err := getRemote(t.user, t.host)
	if err != nil {
		return "", fmt.Errorf("Tunnel get remote client failed: %v", err)
	}
	defer remote.Close()

	session, err := remote.NewSession()
	if err != nil {
		return "", fmt.Errorf("Tunnel get remote session failed: %v", err)
	}
	defer session.Close()

	out, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", fmt.Errorf("Tunnel get port failed: %s, err %v", out, err)
	}

	port := string(out)

	switch port {
	case "404 page not found\n":
		return "", fmt.Errorf("Tunnel weirdness; Thing trying to be its own Mother?; trying again")
	case "no ports available":
		return "", fmt.Errorf("Tunnel no ports available; trying again")
	case "port busy":
		return "", fmt.Errorf("Tunnel port is busy; trying again")
	}

	return port, nil
}

func reverseForward(client net.Conn, remote net.Conn) {
	done := make(chan bool)

	// Start remote -> local data transfer
	go func() {
		io.Copy(client, remote)
		done <- true
	}()

	// Start local -> remote data transfer
	go func() {
		io.Copy(remote, client)
		done <- true
	}()

	<-done
}

func (t *tunnel) tunnel(remotePort string) error {

	// Create an SSH reverse port forwarding tunnel.  Equivalent to:
	//
	//    ssh -NT -R <remotePort>:localhost:<localPort> <user>@<host>
	//

	t.thing.log.printf("Tunnel creating tunnel [ssh -NT -R %s:localhost:%d %s@%s]",
		remotePort, t.portPrivate, t.user, t.host)

	remote, err := getRemote(t.user, t.host)
	if err != nil {
		return fmt.Errorf("Tunnel get remote client failed: %v", err)
	}
	defer remote.Close()

	// Listen on remote server port
	listener, err := remote.Listen("tcp", "localhost:"+remotePort)
	if err != nil {
		return fmt.Errorf("Unable to listen on remote server: %v", err)
	}

	// Handle incoming connections on reverse forwarded tunnel
	address := fmt.Sprintf("localhost:%d", t.portPrivate)
	local, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("Dial into local service error: %v", err)
	}
	defer local.Close()

	client, err := listener.Accept()
	if err != nil {
		return err
	}
	defer client.Close()

	reverseForward(client, local)

	return nil
}

func (t *tunnel) create() {
	var err error
	var port string

	rand.Seed(time.Now().UnixNano())

	for {

		port, err = t.getPort()
		if err != nil {
			t.thing.log.println(err)
			goto again
		}

		t.thing.log.println("Tunnel got port", port)

		err = t.tunnel(port)
		if err != nil {
			t.thing.log.println(err)
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
