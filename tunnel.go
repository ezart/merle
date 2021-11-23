package merle

import (
	"fmt"
	"log"
	"math/rand"
	"os/exec"
	"syscall"
	"time"
)

func (d *Device) tunnelCreate(hubHost, hubUser, hubKey string) {

	var remote string
	var port string

	if hubHost == "" || hubUser == "" || hubKey == "" {
		log.Println("Missing hub connection parameters; skipping tunnel")
		return
	}

	rand.Seed(time.Now().UnixNano())

	for {

		// ssh -i <key> <user>@<hub> curl -s localhost:8080/port?id=xxx

		cmd := exec.Command("ssh", "-i", hubKey,
			hubUser+"@"+hubHost,
			"curl", "-s", "localhost:8080/port?id="+d.id)
		// If the parent process (this app) dies, kill the ssh cmd also
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Pdeathsig: syscall.SIGTERM,
		}
		stdoutStderr, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Get port failed: %s, err %v", stdoutStderr, err)
			goto again
		}
		port = string(stdoutStderr)
		if port == "no ports available" {
			log.Print("No ports available; trying again\n")
			goto again
		}
		if port == "port busy" {
			log.Print("Port is busy; trying again\n")
			goto again
		}

		// ssh -o ExitOnForwardFailure=yes -CNT -i <key> -R 8081:localhost:8080 <hub>
		//
		//  (The ExitOnForwardFailure=yes is to exit ssh if the remote port forwarding fails,
		//   most likely from port already being in-use on the server side).

		remote = fmt.Sprintf("%s:localhost:8080", port)
		log.Println("Creating tunnel:", remote)
		cmd = exec.Command("ssh", "-o", "ExitOnForwardFailure=yes",
			"-CNT", "-i", hubKey,
			"-R", remote, hubUser+"@"+hubHost)
		// If the parent process (this app) dies, kill the ssh cmd also
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Pdeathsig: syscall.SIGTERM,
		}
		stdoutStderr, err = cmd.CombinedOutput()
		if err != nil {
			log.Printf("Create tunnel failed: %s, err %v", stdoutStderr, err)
		}

	again:
		// TODO maybe try some exponential back-off aglo ala TCP

		// Sleep for some number of random seconds between 1 and 10
		// before trying (again).  This will keep us from grinding
		// the CPU trying to connect all the time, and in the case
		// of multi clients starting at exactly the same time will
		// avoid port contention.

		f := rand.Float32() * 10
		log.Printf("Sleeping for %f seconds", f)
		time.Sleep(time.Duration(f*1000) * time.Millisecond)

	}
}
