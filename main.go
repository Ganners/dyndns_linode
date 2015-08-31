package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/Ganners/dyndns_linode/dns_daemon"
	"github.com/Ganners/dyndns_linode/linode_client"
	"github.com/sevlyar/go-daemon"
)

// Launch our application
func main() {

	var configFile string
	var stop bool

	flag.StringVar(
		&configFile, "configFile", "", "The location of your configuration file")
	flag.BoolVar(
		&stop, "stop", false, "Terminate daemon")

	flag.Parse()

	// Add a daemon command to stop, triggered by stop flag
	daemon.AddCommand(daemon.BoolFlag(&stop),
		syscall.SIGTERM, termHandler)

	// Create a context
	cntxt := &daemon.Context{
		PidFileName: "pid",
		PidFilePerm: 0644,
		LogFileName: "dyndns_linode.log",
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
		Args: []string{"dyndns_linode",
			fmt.Sprintf("--configFile=%s", configFile)},
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

	log.Println("Daemon has started, commencing goroutine")

	// Read configuration and get struct
	config, err := readConfiguration(&configFile)
	if err != nil {
		log.Fatal("Could not load configuration file: " + err.Error())
		return
	}
	config.PollRate = 300
	config.API = linode_client.NewAPI(config.ApiKey)

	// Do the work, this gets called again and so everything on the
	// daemon launch which is required should be re-passed through
	// the context
	go dns_daemon.UpdateDaemon(config)

	// Log any signal errors
	err = daemon.ServeSignals()
	if err != nil {
		log.Println("Error:", err)
	}
}

// Handle termination of the daemon
func termHandler(sig os.Signal) error {
	log.Println("Daemon has been terminated")
	return daemon.ErrStop
}

// Reads the configuration and parses the toml
func readConfiguration(file *string) (*dns_daemon.Config, error) {

	bytes, err := ioutil.ReadFile(*file)
	if err != nil {
		return &dns_daemon.Config{}, err
	}

	var conf dns_daemon.Config
	if _, err := toml.Decode(string(bytes), &conf); err != nil {
		return &dns_daemon.Config{}, err
	}

	return &conf, nil
}
