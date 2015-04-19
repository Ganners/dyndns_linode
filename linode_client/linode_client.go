package linode_client

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type DomainList struct {
}

type API struct {
	APIKey string
}

func (api *API) call(action string) (string, error) {

	var requestUri = fmt.Sprintf("https://api.linode.com/?api_key=%s&api_action=%s",
		api.APIKey, action)

	resp, err := http.Get(requestUri)
	if err != nil {
		return "", err
	}

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(contents), nil
}

// Retrieves a domain list
func (api *API) DomainList() (DomainList, error) {

	response, _ := api.call("domain.list")
	log.Println(response)

	return DomainList{}, nil
}

// Updates a given domain
func (api *API) DomainUpdate() error {

	return nil
}

// Constructs a new API
func NewAPI(apiKey string) *API {

	return &API{
		APIKey: apiKey}
}
