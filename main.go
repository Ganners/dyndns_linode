package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"syscall"
	"time"

	"dyndns_linode/linode_client"

	"github.com/sevlyar/go-daemon"
)

type LinodeConfig struct {
	APIKey   string
	Domain   string
	PollRate int16
	API      linode_client.API
}

// Gets the current record from Linode
func getCurrentRecordIP(config *LinodeConfig) string {

	domainList, err := config.API.DomainList()
	if err != nil {

		log.Println(err)
	}

	log.Println(domainList)

	return "IP ADDRESS"
}

func getExternalIP() (string, error) {

	resp, err := http.Get("http://echoip.com")

	if err != nil {
		log.Println("Failed to get public IP", err)
		return "", err
	}

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to get public IP", err)
		return "", err
	}

	defer resp.Body.Close()
	return string(contents), nil
}

// Updates the DNS at scheduled intervals
func DNSUpdateDaemon(config *LinodeConfig) {

	for {

		// Get the current IP address
		ip, err := getExternalIP()

		if err == nil {

			// Check if current DNS record matches
			currentIP := getCurrentRecordIP(config)

			if ip != currentIP {
				// If it doesn't match, update it

			}

			// Sleep for the interval period
			time.Sleep(time.Second * 30)
		}
	}
}

func main() {

	var stop = flag.Bool("stop", false, "Terminate daemon")
	var apiKey = flag.String("apikey", "", "Your Linode API Key")
	var domain = flag.String("domain", "", "Your Linode domain name")
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
	go DNSUpdateDaemon(&LinodeConfig{
		APIKey:   *apiKey,
		Domain:   *domain,
		PollRate: 300,
		API:      *linode_client.NewAPI(*apiKey)})

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
