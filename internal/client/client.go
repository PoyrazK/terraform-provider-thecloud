package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	errUnexpectedStatus = "unexpected status code: %d"
)

// Client is the base structure for interacting with The Cloud API
type Client struct {
	Endpoint   string
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates a new API client for The Cloud
func NewClient(endpoint, apiKey string) *Client {
	return &Client{
		Endpoint: endpoint,
		APIKey:   apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// DoRequest performs an HTTP request with the necessary headers
func (c *Client) DoRequest(req *http.Request) (*http.Response, error) {
	req.Header.Set("X-API-Key", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	return c.HTTPClient.Do(req)
}

func (c *Client) BuildURL(path string) string {
	return fmt.Sprintf("%s%s", c.Endpoint, path)
}

// VPC represents the API response for a VPC
type VPC struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CIDRBlock string `json:"cidr_block"`
	Status    string `json:"status"`
}

func (c *Client) CreateVPC(name, cidr string) (*VPC, error) {
	payload := map[string]string{
		"name":       name,
		"cidr_block": cidr,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", c.BuildURL("/vpcs"), bytes.NewBuffer(body))
	resp, err := c.DoRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf(errUnexpectedStatus, resp.StatusCode)
	}

	var vpc VPC
	if err := json.NewDecoder(resp.Body).Decode(&vpc); err != nil {
		return nil, err
	}

	return &vpc, nil
}

func (c *Client) GetVPC(id string) (*VPC, error) {
	req, _ := http.NewRequest("GET", c.BuildURL(fmt.Sprintf("/vpcs/%s", id)), nil)
	resp, err := c.DoRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // Not found
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(errUnexpectedStatus, resp.StatusCode)
	}

	var vpc VPC
	if err := json.NewDecoder(resp.Body).Decode(&vpc); err != nil {
		return nil, err
	}

	return &vpc, nil
}

func (c *Client) DeleteVPC(id string) error {
	req, _ := http.NewRequest("DELETE", c.BuildURL(fmt.Sprintf("/vpcs/%s", id)), nil)
	resp, err := c.DoRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf(errUnexpectedStatus, resp.StatusCode)
	}

	return nil
}

// Instance represents the API response for an Instance
type Instance struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Image     string `json:"image"`
	Ports     string `json:"ports"`
	VpcID     string `json:"vpc_id"`
	Status    string `json:"status"`
	IPAddress string `json:"ip_address"`
}

type LaunchInstanceRequest struct {
	Name     string `json:"name"`
	Image    string `json:"image"`
	Ports    string `json:"ports"`
	VpcID    string `json:"vpc_id"`
	SubnetID string `json:"subnet_id"`
}

func (c *Client) CreateInstance(reqBody LaunchInstanceRequest) (*Instance, error) {
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", c.BuildURL("/instances"), bytes.NewBuffer(body))
	resp, err := c.DoRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf(errUnexpectedStatus, resp.StatusCode)
	}

	var instance Instance
	if err := json.NewDecoder(resp.Body).Decode(&instance); err != nil {
		return nil, err
	}

	return &instance, nil
}

func (c *Client) GetInstance(id string) (*Instance, error) {
	req, _ := http.NewRequest("GET", c.BuildURL(fmt.Sprintf("/instances/%s", id)), nil)
	resp, err := c.DoRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // Not found
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(errUnexpectedStatus, resp.StatusCode)
	}

	var instance Instance
	if err := json.NewDecoder(resp.Body).Decode(&instance); err != nil {
		return nil, err
	}

	return &instance, nil
}

func (c *Client) DeleteInstance(id string) error {
	req, _ := http.NewRequest("DELETE", c.BuildURL(fmt.Sprintf("/instances/%s", id)), nil)
	resp, err := c.DoRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf(errUnexpectedStatus, resp.StatusCode)
	}

	return nil
}
