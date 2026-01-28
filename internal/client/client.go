package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	errUnexpectedStatus = "unexpected status code: %d"
)

// APIError represents the structured error from the API
type APIError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("[%s] %s (code: %s)", e.Type, e.Message, e.Code)
}

// APIResponse wraps the standard API response structure
type APIResponse struct {
	Data  json.RawMessage `json:"data,omitempty"`
	Error *APIError       `json:"error,omitempty"`
}

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

func (c *Client) do(ctx context.Context, method, path string, body interface{}, v interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BuildURL(path), bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return resp, nil
	}

	if resp.StatusCode >= 400 {
		return resp, c.handleError(resp)
	}

	if v != nil {
		if err := c.decodeResponse(resp, v); err != nil {
			return resp, err
		}
	}

	return resp, nil
}

func (c *Client) handleError(resp *http.Response) error {
	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err == nil && apiResp.Error != nil {
		return apiResp.Error
	}
	return fmt.Errorf(errUnexpectedStatus, resp.StatusCode)
}

func (c *Client) decodeResponse(resp *http.Response, v interface{}) error {
	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Error != nil {
		return apiResp.Error
	}

	if apiResp.Data != nil && v != nil {
		if err := json.Unmarshal(apiResp.Data, v); err != nil {
			return fmt.Errorf("failed to unmarshal data: %w", err)
		}
	}

	return nil
}

// VPC represents the API response for a VPC
type VPC struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CIDRBlock string `json:"cidr_block"`
	Status    string `json:"status"`
}

func (c *Client) CreateVPC(ctx context.Context, name, cidr string) (*VPC, error) {
	payload := map[string]string{
		"name":       name,
		"cidr_block": cidr,
	}

	var vpc VPC
	_, err := c.do(ctx, "POST", "/vpcs", payload, &vpc)
	if err != nil {
		return nil, err
	}

	return &vpc, nil
}

func (c *Client) GetVPC(ctx context.Context, id string) (*VPC, error) {
	var vpc VPC
	resp, err := c.do(ctx, "GET", fmt.Sprintf("/vpcs/%s", id), nil, &vpc)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	return &vpc, nil
}

func (c *Client) DeleteVPC(ctx context.Context, id string) error {
	_, err := c.do(ctx, "DELETE", fmt.Sprintf("/vpcs/%s", id), nil, nil)
	return err
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

func (c *Client) CreateInstance(ctx context.Context, reqBody LaunchInstanceRequest) (*Instance, error) {
	var instance Instance
	_, err := c.do(ctx, "POST", "/instances", reqBody, &instance)
	if err != nil {
		return nil, err
	}

	return &instance, nil
}

func (c *Client) GetInstance(ctx context.Context, id string) (*Instance, error) {
	var instance Instance
	resp, err := c.do(ctx, "GET", fmt.Sprintf("/instances/%s", id), nil, &instance)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	return &instance, nil
}

func (c *Client) DeleteInstance(ctx context.Context, id string) error {
	_, err := c.do(ctx, "DELETE", fmt.Sprintf("/instances/%s", id), nil, nil)
	return err
}

// Volume represents the API response for a Volume
type Volume struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	SizeGB int    `json:"size_gb"`
	Status string `json:"status"`
}

func (c *Client) CreateVolume(ctx context.Context, name string, sizeGB int) (*Volume, error) {
	payload := map[string]interface{}{
		"name":    name,
		"size_gb": sizeGB,
	}

	var vol Volume
	_, err := c.do(ctx, "POST", "/volumes", payload, &vol)
	if err != nil {
		return nil, err
	}

	return &vol, nil
}

func (c *Client) GetVolume(ctx context.Context, id string) (*Volume, error) {
	var vol Volume
	resp, err := c.do(ctx, "GET", fmt.Sprintf("/volumes/%s", id), nil, &vol)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	return &vol, nil
}

func (c *Client) DeleteVolume(ctx context.Context, id string) error {
	_, err := c.do(ctx, "DELETE", fmt.Sprintf("/volumes/%s", id), nil, nil)
	return err
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

func (c *Client) CreateSecurityGroup(ctx context.Context, vpcID, name, description string) (*SecurityGroup, error) {
	payload := map[string]string{
		"vpc_id":      vpcID,
		"name":        name,
		"description": description,
	}

	var sg SecurityGroup
	_, err := c.do(ctx, "POST", "/security-groups", payload, &sg)
	if err != nil {
		return nil, err
	}

	return &sg, nil
}

func (c *Client) GetSecurityGroup(ctx context.Context, id string) (*SecurityGroup, error) {
	var sg SecurityGroup
	resp, err := c.do(ctx, "GET", fmt.Sprintf("/security-groups/%s", id), nil, &sg)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	return &sg, nil
}

func (c *Client) DeleteSecurityGroup(ctx context.Context, id string) error {
	_, err := c.do(ctx, "DELETE", fmt.Sprintf("/security-groups/%s", id), nil, nil)
	return err
}

func (c *Client) AddSecurityRule(ctx context.Context, groupID string, rule SecurityRule) (*SecurityRule, error) {
	var result SecurityRule
	_, err := c.do(ctx, "POST", fmt.Sprintf("/security-groups/%s/rules", groupID), rule, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) RemoveSecurityRule(ctx context.Context, ruleID string) error {
	_, err := c.do(ctx, "DELETE", fmt.Sprintf("/security-groups/rules/%s", ruleID), nil, nil)
	return err
}

// LoadBalancer represents the API response for a Load Balancer
type LoadBalancer struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	VpcID     string     `json:"vpc_id"`
	Port      int        `json:"port"`
	Algorithm string     `json:"algorithm"`
	Status    string     `json:"status"`
	Targets   []LBTarget `json:"targets,omitempty"`
}

// LBTarget represents a target within a Load Balancer
type LBTarget struct {
	InstanceID string `json:"instance_id"`
	Port       int    `json:"port"`
	Weight     int    `json:"weight"`
}

func (c *Client) CreateLoadBalancer(ctx context.Context, name, vpcID string, port int, algorithm string) (*LoadBalancer, error) {
	payload := map[string]interface{}{
		"name":      name,
		"vpc_id":    vpcID,
		"port":      port,
		"algorithm": algorithm,
	}

	var lb LoadBalancer
	_, err := c.do(ctx, "POST", "/lb", payload, &lb)
	if err != nil {
		return nil, err
	}

	return &lb, nil
}

func (c *Client) GetLoadBalancer(ctx context.Context, id string) (*LoadBalancer, error) {
	var lb LoadBalancer
	resp, err := c.do(ctx, "GET", fmt.Sprintf("/lb/%s", id), nil, &lb)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	// Also fetch targets
	targets, _ := c.ListLBTargets(ctx, id)
	lb.Targets = targets

	return &lb, nil
}

func (c *Client) DeleteLoadBalancer(ctx context.Context, id string) error {
	_, err := c.do(ctx, "DELETE", fmt.Sprintf("/lb/%s", id), nil, nil)
	return err
}

func (c *Client) AddLBTarget(ctx context.Context, lbID string, target LBTarget) error {
	_, err := c.do(ctx, "POST", fmt.Sprintf("/lb/%s/targets", lbID), target, nil)
	return err
}

func (c *Client) RemoveLBTarget(ctx context.Context, lbID, instanceID string) error {
	_, err := c.do(ctx, "DELETE", fmt.Sprintf("/lb/%s/targets/%s", lbID, instanceID), nil, nil)
	return err
}

func (c *Client) ListLBTargets(ctx context.Context, lbID string) ([]LBTarget, error) {
	var targets []LBTarget
	_, err := c.do(ctx, "GET", fmt.Sprintf("/lb/%s/targets", lbID), nil, &targets)
	if err != nil {
		return nil, err
	}
	return targets, nil
}

// Secret represents the API response for a Secret
type Secret struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Value       string `json:"value,omitempty"`
	Description string `json:"description"`
}

func (c *Client) CreateSecret(ctx context.Context, name, value, description string) (*Secret, error) {
	payload := map[string]string{
		"name":        name,
		"value":       value,
		"description": description,
	}

	var secret Secret
	_, err := c.do(ctx, "POST", "/secrets", payload, &secret)
	if err != nil {
		return nil, err
	}

	return &secret, nil
}

func (c *Client) GetSecret(ctx context.Context, id string) (*Secret, error) {
	var secret Secret
	resp, err := c.do(ctx, "GET", fmt.Sprintf("/secrets/%s", id), nil, &secret)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	return &secret, nil
}

func (c *Client) DeleteSecret(ctx context.Context, id string) error {
	_, err := c.do(ctx, "DELETE", fmt.Sprintf("/secrets/%s", id), nil, nil)
	return err
}

// APIKey represents the API response for an API Key
type APIKey struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Key       string `json:"key,omitempty"`
	CreatedAt string `json:"created_at"`
}

func (c *Client) CreateAPIKey(ctx context.Context, name string) (*APIKey, error) {
	payload := map[string]string{
		"name": name,
	}

	var key APIKey
	_, err := c.do(ctx, "POST", "/auth/keys", payload, &key)
	if err != nil {
		return nil, err
	}

	return &key, nil
}

func (c *Client) ListAPIKeys(ctx context.Context) ([]APIKey, error) {
	var keys []APIKey
	_, err := c.do(ctx, "GET", "/auth/keys", nil, &keys)
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (c *Client) RevokeAPIKey(ctx context.Context, id string) error {
	_, err := c.do(ctx, "DELETE", fmt.Sprintf("/auth/keys/%s", id), nil, nil)
	return err
}

// ScalingGroup represents the API response for an Auto-Scaling Group
type ScalingGroup struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	VpcID          string `json:"vpc_id"`
	LoadBalancerID string `json:"load_balancer_id,omitempty"`
	Image          string `json:"image"`
	Ports          string `json:"ports"`
	MinInstances   int    `json:"min_instances"`
	MaxInstances   int    `json:"max_instances"`
	DesiredCount   int    `json:"desired_count"`
	Status         string `json:"status"`
}

func (c *Client) CreateScalingGroup(ctx context.Context, params map[string]interface{}) (*ScalingGroup, error) {
	var group ScalingGroup
	_, err := c.do(ctx, "POST", "/autoscaling/groups", params, &group)
	if err != nil {
		return nil, err
	}

	return &group, nil
}

func (c *Client) GetScalingGroup(ctx context.Context, id string) (*ScalingGroup, error) {
	var group ScalingGroup
	resp, err := c.do(ctx, "GET", fmt.Sprintf("/autoscaling/groups/%s", id), nil, &group)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	return &group, nil
}

func (c *Client) DeleteScalingGroup(ctx context.Context, id string) error {
	_, err := c.do(ctx, "DELETE", fmt.Sprintf("/autoscaling/groups/%s", id), nil, nil)
	return err
}

func (c *Client) ListVPCs(ctx context.Context) ([]VPC, error) {
	var vpcs []VPC
	_, err := c.do(ctx, "GET", "/vpcs", nil, &vpcs)
	if err != nil {
		return nil, err
	}
	return vpcs, nil
}

func (c *Client) ListInstances(ctx context.Context) ([]Instance, error) {
	var instances []Instance
	_, err := c.do(ctx, "GET", "/instances", nil, &instances)
	if err != nil {
		return nil, err
	}
	return instances, nil
}

func (c *Client) ListVolumes(ctx context.Context) ([]Volume, error) {
	var volumes []Volume
	_, err := c.do(ctx, "GET", "/volumes", nil, &volumes)
	if err != nil {
		return nil, err
	}
	return volumes, nil
}
