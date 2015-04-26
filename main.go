package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/Ganners/dyndns_linode/linode_client"
	"github.com/sevlyar/go-daemon"
)

// Holds our Linode and application configuration
type Config struct {
	APIKey    string
	Domain    string
	SubDomain string
	PollRate  time.Duration
	API       linode_client.API
}

// Stores the resource information that we need to hold
// on to
type ResourceInfo struct {
	ResourceID int64
	DomainID   int
	ResourceIP string
}

// Gets the current record from Linode
func fetchResourceInfo(config *Config) (ResourceInfo, error) {

	domainList, err := config.API.DomainList()
	if err != nil {
		return ResourceInfo{}, err
	}

	// Loop the list of domains to hunt for the one we want
	for _, domainData := range domainList.Data {
		if domainData.Domain == config.Domain {

			// Grab the Domain ID
			domainID := domainData.Domainid

			// Grab the resource ID
			resources, err := config.API.DomainResourceList(domainID)
			if err != nil {
				return ResourceInfo{}, err
			}

			// Loop through the resources until we find our subdomain
			for _, domainResourceData := range resources.Data {

				// Hunt for our subdomain
				if domainResourceData.Name == config.SubDomain {
					return ResourceInfo{
							domainResourceData.Resourceid,
							domainResourceData.Domainid,
							domainResourceData.Target},
						nil
				}
			}
		}
	}

	return ResourceInfo{}, err
}

// Grabs our external IP, requires echoip.com to be
// online but it seems fairly reliable
func getExternalIP() (string, error) {

	// Add a 5 second timeout
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get("http://www.echoip.com")

	if err != nil {
		log.Println("Failed to get public IP: ", err)
		return "", err
	}

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to get public IP: ", err)
		return "", err
	}

	defer resp.Body.Close()
	return string(contents), nil
}

// Updates the DNS at scheduled intervals
func DNSUpdateDaemon(config *Config) {

	for {

		log.Println("Commencing DNS Update")
		// Get the current IP address
		ip, err := getExternalIP()

		if err == nil {

			// Check if current DNS record matches
			resourceInfo, err := fetchResourceInfo(config)
			if err != nil {
				log.Println("Error: ", err)
			} else {

				if ip != resourceInfo.ResourceIP {
					// If it doesn't match, update it

					err := config.API.DomainResourceUpdate(
						int(resourceInfo.ResourceID),
						resourceInfo.DomainID,
						ip)

					if err != nil {
						log.Println("Error: ", err)
					}
				}
			}
		}

		// Sleep for the interval period
		time.Sleep(time.Second * (30 * 5))
	}
}

// Handle termination of the daemon
func termHandler(sig os.Signal) error {
	log.Println("Daemon has been terminated")
	return daemon.ErrStop
}

// Launch our application
func main() {

	var apiKey string
	var domain string
	var subDomain string
	flag.StringVar(&apiKey, "apikey", "", "Your Linode API Key")
	flag.StringVar(&domain, "domain", "", "Your Linode domain name")
	flag.StringVar(&subDomain, "subdomain", "", "Your Linode domain name")

	var stop = flag.Bool("stop", false, "Terminate daemon")
	flag.Parse()

	// Add a daemon command to stop, triggered by stop flag
	daemon.AddCommand(daemon.BoolFlag(stop),
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
			fmt.Sprintf("--apikey=%s", apiKey),
			fmt.Sprintf("--domain=%s", domain),
			fmt.Sprintf("--subdomain=%s", subDomain)},
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

	// Do the work, this gets called again and so everything on the
	// daemon launch which is required should be re-passed through
	// the context
	go DNSUpdateDaemon(&Config{
		APIKey:    apiKey,
		Domain:    domain,
		SubDomain: subDomain,
		PollRate:  300,
		API:       *linode_client.NewAPI(apiKey)})

	// Log any signal errors
	err = daemon.ServeSignals()
	if err != nil {
		log.Println("Error:", err)
	}
}
