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

// Volume represents the API response for a Volume
type Volume struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	SizeGB int    `json:"size_gb"`
	Status string `json:"status"`
}

func (c *Client) CreateVolume(name string, sizeGB int) (*Volume, error) {
	payload := map[string]interface{}{
		"name":    name,
		"size_gb": sizeGB,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", c.BuildURL("/volumes"), bytes.NewBuffer(body))
	resp, err := c.DoRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf(errUnexpectedStatus, resp.StatusCode)
	}

	var vol Volume
	if err := json.NewDecoder(resp.Body).Decode(&vol); err != nil {
		return nil, err
	}

	return &vol, nil
}

func (c *Client) GetVolume(id string) (*Volume, error) {
	req, _ := http.NewRequest("GET", c.BuildURL(fmt.Sprintf("/volumes/%s", id)), nil)
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

	var vol Volume
	if err := json.NewDecoder(resp.Body).Decode(&vol); err != nil {
		return nil, err
	}

	return &vol, nil
}

func (c *Client) DeleteVolume(id string) error {
	req, _ := http.NewRequest("DELETE", c.BuildURL(fmt.Sprintf("/volumes/%s", id)), nil)
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

// SecurityGroup represents the API response for a Security Group
type SecurityGroup struct {
	ID          string         `json:"id"`
	VPCID       string         `json:"vpc_id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Rules       []SecurityRule `json:"rules,omitempty"`
}

// SecurityRule represents a rule within a Security Group
type SecurityRule struct {
	ID        string `json:"id"`
	GroupID   string `json:"group_id"`
	Direction string `json:"direction"`
	Protocol  string `json:"protocol"`
	PortMin   int    `json:"port_min,omitempty"`
	PortMax   int    `json:"port_max,omitempty"`
	CIDR      string `json:"cidr"`
	Priority  int    `json:"priority"`
}

func (c *Client) CreateSecurityGroup(vpcID, name, description string) (*SecurityGroup, error) {
	payload := map[string]string{
		"vpc_id":      vpcID,
		"name":        name,
		"description": description,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", c.BuildURL("/security-groups"), bytes.NewBuffer(body))
	resp, err := c.DoRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf(errUnexpectedStatus, resp.StatusCode)
	}

	var sg SecurityGroup
	if err := json.NewDecoder(resp.Body).Decode(&sg); err != nil {
		return nil, err
	}

	return &sg, nil
}

func (c *Client) GetSecurityGroup(id string) (*SecurityGroup, error) {
	req, _ := http.NewRequest("GET", c.BuildURL(fmt.Sprintf("/security-groups/%s", id)), nil)
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

	var sg SecurityGroup
	if err := json.NewDecoder(resp.Body).Decode(&sg); err != nil {
		return nil, err
	}

	return &sg, nil
}

func (c *Client) DeleteSecurityGroup(id string) error {
	req, _ := http.NewRequest("DELETE", c.BuildURL(fmt.Sprintf("/security-groups/%s", id)), nil)
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

func (c *Client) AddSecurityRule(groupID string, rule SecurityRule) (*SecurityRule, error) {
	body, _ := json.Marshal(rule)

	req, _ := http.NewRequest("POST", c.BuildURL(fmt.Sprintf("/security-groups/%s/rules", groupID)), bytes.NewBuffer(body))
	resp, err := c.DoRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf(errUnexpectedStatus, resp.StatusCode)
	}

	var result SecurityRule
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) RemoveSecurityRule(ruleID string) error {
	req, _ := http.NewRequest("DELETE", c.BuildURL(fmt.Sprintf("/security-groups/rules/%s", ruleID)), nil)
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
