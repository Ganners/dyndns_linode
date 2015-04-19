package main

import (
	"flag"
	"log"
	"os"
	"syscall"

	"github.com/sevlyar/go-daemon"
)

type LinodeConfig struct {
}

// Updates the DNS at scheduled intervals
func DNSUpdateDaemon(config *LinodeConfig) {

	// Get the current IP address

	for {

		// Check if current DNS record matches

		// If it doesn't match, update it
	}
}

func main() {

	var stop = flag.Bool("stop", false, "Terminate")
	flag.Parse()

	// Add a daemon command to stop, triggered by stop flag
	daemon.AddCommand(daemon.BoolFlag(stop),
		syscall.SIGTERM, termHandler)

	// Create a context
	cntxt := &daemon.Context{
		PidFileName: "pid",
		PidFilePerm: 0644,
		LogFileName: "log",
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
		Args:        []string{"dyndns_linode"},
	}

	// Send commands if flags were sent
	if len(daemon.ActiveFlags()) > 0 {
		d, err := cntxt.Search()
		if err != nil {
			log.Fatalln("Unable send signal to the daemon:", err)
		}
		daemon.SendCommands(d)
		return
	}

	// Create the daemon
	d, err := cntxt.Reborn()
	if err != nil {
		log.Fatalln(err)
	}
	if d != nil {
		return
	}
	defer cntxt.Release()

	// Do the work
	go DNSUpdateDaemon(&LinodeConfig{})

	// Log any signal errors
	err = daemon.ServeSignals()
	if err != nil {
		log.Println("Error:", err)
	}
}

// Handle termination of the daemon
func termHandler(sig os.Signal) error {
	log.Println("Terminating")
	return daemon.ErrStop
}
