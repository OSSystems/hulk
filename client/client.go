package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/OSSystems/hulk/api/types"
)

// Client is the API client
type Client struct {
	address string
	proto   string
	host    string
	client  *http.Client
}

// NewEnvClient initializes a new API client based on environment variable
func NewEnvClient() (*Client, error) {
	return NewClient(os.Getenv("HULK_ADDRESS"))
}

// NewClient initializes a new API client for the given address
func NewClient(address string) (*Client, error) {
	parts := strings.SplitN(address, "://", 2)

	if len(parts) != 2 {
		return nil, errors.New("Invalid address")
	}

	proto := parts[0]
	host := parts[1]

	client := &http.Client{}

	if proto == "tcp" {
		client.Transport = &http.Transport{}
	} else if proto == "unix" {
		client.Transport = NewUnixTransport(host)
	}

	return &Client{
		address: address,
		proto:   proto,
		host:    host,
		client:  client,
	}, nil
}

func (cli *Client) buildRequest(method, path string, body io.Reader) (*http.Request, error) {
	host := cli.host

	if cli.proto == "unix" {
		host = ""
	}

	url := fmt.Sprintf("%s://%s%s", cli.proto, host, path)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	return req, nil
}

func (cli *Client) sendRequest(method, path string, body io.Reader) (*http.Response, error) {
	req, err := cli.buildRequest(method, path, body)
	if err != nil {
		return nil, err
	}

	resp, err := cli.client.Do(req)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error status code %d", resp.StatusCode)
	}

	return resp, err
}

// ServiceList returns the list of services in the Hulk Daemon
func (cli *Client) ServiceList() ([]*types.Service, error) {
	resp, err := cli.sendRequest("GET", "/services", nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var services []*types.Service

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &services)

	return services, err
}

// GetService gets service details from Hulk Daemon
func (cli *Client) GetService(name string) (*types.Service, error) {
	resp, err := cli.sendRequest("GET", fmt.Sprintf("/services/%s", name), nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var service *types.Service

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &service)

	return service, err
}
