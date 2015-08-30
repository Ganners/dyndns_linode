package dns_daemon

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/Ganners/dyndns_linode/linode_client"
)

// Holds the configuration
type Config struct {
	ApiKey   string
	PollRate int
	API      *linode_client.API
	Domains  []struct {
		Domain    string
		Subdomain string
		ResInf    *ResourceInfo
	}
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

	for _, domain := range c.Domains {
		if domain.Subdomain == searchName {
			return true, domain.ResInf
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
					*resInf = *&ResourceInfo{
						domainResourceData.Resourceid,
						domainResourceData.Domainid,
						domainResourceData.Target}
				}
			}
			return nil
		}
	}

	return err
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
func UpdateDaemon(config *Config) {

	for {

		log.Println("Commencing DNS Update")

		// Get the current IP address
		ip, err := getExternalIP()

		if err != nil {
			log.Println("Error: ", err)
			continue
		}

		log.Println("IP address retrieved successfully, continuing")

		// Check if current DNS record matches
		err = populateResourceInfo(config)
		if err != nil {
			log.Println("Error: ", err)
			continue
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

		// Sleep for the interval period
		time.Sleep(time.Second * (30 * 5))
	}
}
