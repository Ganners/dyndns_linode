package dns_daemon

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Ganners/dyndns_linode/linode_client"
)

// Hold domains and domain data
type Domain struct {
	Domain    string `toml:"domain"`
	Subdomain string `toml:"subdomain"`
	ResInf    *ResourceInfo
}

// Holds the configuration
type Config struct {
	ApiKey   string `toml:"api_key"`
	PollRate int    `toml:"poll_rate"`
	API      *linode_client.API
	Domains  []Domain `toml:"domain"`
}

// Returns bool if the config domains list contains this domain
func (c *Config) hasDomain(searchDomain string) bool {

	for _, domain := range c.Domains {
		if domain.Domain == searchDomain {
			return true
		}
	}
	return false
}

// Returns bool if the config domains list contains this domain
func (c *Config) hasSubdomain(searchName string) (bool, *ResourceInfo) {

	for i, domain := range c.Domains {
		if domain.Subdomain == searchName {
			c.Domains[i].ResInf = &ResourceInfo{}
			return true, c.Domains[i].ResInf
		}
	}
	return false, nil
}

// Stores the resource information that we need to hold
// on to
type ResourceInfo struct {
	ResourceID int64
	DomainID   int
	ResourceIP string
}

// Loops through the configuration domain list and populates the
// resource info if it is currently empty
func populateResourceInfo(config *Config) error {

	// Grab the domain list
	domainList, err := config.API.DomainList()
	if err != nil {
		return err
	}

	// Loop the list of domains to hunt for the one we want
	for _, domainData := range domainList.Data {

		// Is this a domain we care about?
		if config.hasDomain(domainData.Domain) {

			// Grab the Domain ID
			domainID := domainData.Domainid

			// Grab the resource ID
			resources, err := config.API.DomainResourceList(domainID)
			if err != nil {
				return err
			}

			// Loop through the resources until we find our subdomain
			for _, domainResourceData := range resources.Data {

				// Hunt for our subdomain and update (set pointer to new resource)
				if found, resInf := config.hasSubdomain(domainResourceData.Name); found {

					// Update pointer
					resInf.ResourceID = domainResourceData.Resourceid
					resInf.DomainID = domainResourceData.Domainid
					resInf.ResourceIP = domainResourceData.Target
				}
			}
			return nil
		}
	}

	return err
}

// Grabs our external IP, requires ifconfig.co to be
// online but it seems fairly reliable
func getExternalIP() (string, error) {

	// Add a 5 second timeout
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get("https://ifconfig.co/")
	if err != nil {
		log.Println("Failed to get public IP: ", err)
		return "", err
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to get public IP: ", err)
		return "", err
	}

	ip := strings.TrimSpace(string(contents))

	log.Println("Retrieved public IP: ", ip)

	return ip, nil
}

// Updates the DNS at scheduled intervals
func UpdateDaemon(config *Config) {

	for {

		log.Println("Commencing DNS Update")

		// Get the current IP address
		ip, err := getExternalIP()

		if err != nil {
			log.Println("Error: ", err)
			goto Retry
		}

		log.Println("IP address retrieved successfully, continuing")

		// Check if current DNS record matches
		err = populateResourceInfo(config)
		if err != nil {
			log.Println("Error: ", err)
			goto Retry
		}

		log.Println("Successfully gathered resource info")

		// Range domains and check
		for _, domain := range config.Domains {
			if ip != domain.ResInf.ResourceIP {
				// If it doesn't match, update it

				err := config.API.DomainResourceUpdate(
					int(domain.ResInf.ResourceID),
					domain.ResInf.DomainID,
					ip)

				if err != nil {
					log.Println("Error: ", err)
				} else {
					log.Println("Successfully updated IP")
				}
			} else {
				log.Println("IP address not changed, skipping update")
			}
		}

Retry:
		// Sleep for the interval period
		time.Sleep(time.Second * (60 * 5))
	}
}
