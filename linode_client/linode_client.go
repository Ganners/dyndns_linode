package linode_client

import (
	"encoding/json"
	"fmt"
	"log"
	"io/ioutil"
	"net/http"
	"strconv"
)

type DomainList struct {
	Action string `json:"ACTION"`
	Data   []struct {
		Axfrips         string `json:"AXFR_IPS"`
		Description     string `json:"DESCRIPTION"`
		Domain          string `json:"DOMAIN"`
		Domainid        int    `json:"DOMAINID"`
		Expiresec       int64  `json:"EXPIRE_SEC"`
		Lpmdisplaygroup string `json:"LPM_DISPLAYGROUP"`
		Masterips       string `json:"MASTER_IPS"`
		Refreshsec      int64  `json:"REFRESH_SEC"`
		Retrysec        int64  `json:"RETRY_SEC"`
		Soaemail        string `json:"SOA_EMAIL"`
		Status          int64  `json:"STATUS"`
		Ttlsec          int64  `json:"TTL_SEC"`
		Type            string `json:"TYPE"`
	} `json:"DATA"`
	Errorarray []interface{} `json:"ERRORARRAY"`
}

type DomainResourceList struct {
	Action string `json:"ACTION"`
	Data   []struct {
		Domainid   int    `json:"DOMAINID"`
		Name       string `json:"NAME"`
		Port       int64  `json:"PORT"`
		Priority   int64  `json:"PRIORITY"`
		Protocol   string `json:"PROTOCOL"`
		Resourceid int64  `json:"RESOURCEID"`
		Target     string `json:"TARGET"`
		Ttlsec     int64  `json:"TTL_SEC"`
		Type       string `json:"TYPE"`
		Weight     int64  `json:"WEIGHT"`
	} `json:"DATA"`
	Errorarray []interface{} `json:"ERRORARRAY"`
}

type DomainResourceUpdate struct {
	Action string `json:"ACTION"`
	Data   struct {
		Resourceid float64 `json:"ResourceID"`
	} `json:"DATA"`
	Errorarray []interface{} `json:"ERRORARRAY"`
}

type API struct {
	APIKey string
}

// Calls the API and returns some bytes
func (api *API) call(action string, params []string) ([]byte, error) {

	// Combine arguments
	var args string
	for _, param := range params {
		args += fmt.Sprintf("&%s", param)
	}

	var requestUri = fmt.Sprintf(
		"https://api.linode.com/?api_key=%s&api_action=%s%s",
		api.APIKey, action, args)

	resp, err := http.Get(requestUri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return contents, nil
}

// Retrieves a domain list
func (api *API) DomainList() (DomainList, error) {

	response, _ := api.call("domain.list", nil)
	var domainList DomainList
	err := json.Unmarshal(response, &domainList)

	if err != nil {
		log.Println(string(response))
		return DomainList{}, err
	}

	return domainList, nil
}

// Grabs a list of resources (records) for a domain
func (api *API) DomainResourceList(domainId int) (DomainResourceList, error) {

	response, _ := api.call("domain.resource.list",
		[]string{fmt.Sprintf("DOMAINID=%s", strconv.Itoa(domainId))})

	var domainResourceList DomainResourceList
	err := json.Unmarshal(response, &domainResourceList)

	if err != nil {
		log.Println(string(response))
		return DomainResourceList{}, err
	}

	return domainResourceList, nil
}

// Updates a resource with a target (i.e. updates subdomain with new ip)
func (api *API) DomainResourceUpdate(
	domainResourceId int, domainId int, newTarget string) error {

	response, _ := api.call("domain.resource.update",
		[]string{
			fmt.Sprintf("RESOURCEID=%s", strconv.Itoa(domainResourceId)),
			fmt.Sprintf("DOMAINID=%s", strconv.Itoa(domainId)),
			fmt.Sprintf("TARGET=%s", newTarget)})

	var updateResponse DomainResourceUpdate
	err := json.Unmarshal(response, &updateResponse)

	if err != nil {
		log.Println(string(response))
		return err
	}

	return nil
}

// Constructs a new API
func NewAPI(apiKey string) *API {

	return &API{
		APIKey: apiKey}
}
