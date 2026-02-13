package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
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
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5
	retryClient.RetryWaitMin = 1 * time.Second
	retryClient.RetryWaitMax = 30 * time.Second
	retryClient.Logger = nil

	return &Client{
		Endpoint:   endpoint,
		APIKey:     apiKey,
		HTTPClient: retryClient.StandardClient(),
	}
}

func (c *Client) BuildURL(path string) string {
	return fmt.Sprintf("%s%s", c.Endpoint, path)
}

func (c *Client) do(ctx context.Context, method, path string, body interface{}, v interface{}) (int, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return 0, err
		}
		bodyReader = bytes.NewBuffer(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BuildURL(path), bodyReader)
	if err != nil {
		return 0, err
	}

	req.Header.Set("X-API-Key", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close() // nolint:errcheck

	if resp.StatusCode == http.StatusNotFound {
		return resp.StatusCode, nil
	}

	if resp.StatusCode >= 400 {
		return resp.StatusCode, c.handleError(resp)
	}

	if v != nil {
		if err := c.decodeResponse(resp, v); err != nil {
			return resp.StatusCode, err
		}
	}

	return resp.StatusCode, nil
}

func (c *Client) handleError(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read error body: %w", err)
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	var apiResp struct {
		Error interface{} `json:"error"`
	}
	if err := json.Unmarshal(body, &apiResp); err == nil && apiResp.Error != nil {
		switch v := apiResp.Error.(type) {
		case string:
			return fmt.Errorf("[%d] %s", resp.StatusCode, v)
		case map[string]interface{}:
			if msg, ok := v["message"].(string); ok {
				return fmt.Errorf("[%d] %s", resp.StatusCode, msg)
			}
		}
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
	status, err := c.do(ctx, "GET", fmt.Sprintf("/vpcs/%s", id), nil, &vpc)
	if err != nil {
		return nil, err
	}

	if status == http.StatusNotFound {
		return nil, nil // nolint:nilnil
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
	status, err := c.do(ctx, "GET", fmt.Sprintf("/instances/%s", id), nil, &instance)
	if err != nil {
		return nil, err
	}

	if status == http.StatusNotFound {
		return nil, nil // nolint:nilnil
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
	status, err := c.do(ctx, "GET", fmt.Sprintf("/volumes/%s", id), nil, &vol)
	if err != nil {
		return nil, err
	}

	if status == http.StatusNotFound {
		return nil, nil // nolint:nilnil
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
	ID        string `json:"id,omitempty"`
	GroupID   string `json:"group_id,omitempty"`
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
	status, err := c.do(ctx, "GET", fmt.Sprintf("/security-groups/%s", id), nil, &sg)
	if err != nil {
		return nil, err
	}

	if status == http.StatusNotFound {
		return nil, nil // nolint:nilnil
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
	status, err := c.do(ctx, "GET", fmt.Sprintf("/lb/%s", id), nil, &lb)
	if err != nil {
		return nil, err
	}

	if status == http.StatusNotFound {
		return nil, nil // nolint:nilnil
	}

	// Also fetch targets
	targets, err := c.ListLBTargets(ctx, id)
	if err != nil {
		return nil, err
	}
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
	status, err := c.do(ctx, "GET", fmt.Sprintf("/secrets/%s", id), nil, &secret)
	if err != nil {
		return nil, err
	}

	if status == http.StatusNotFound {
		return nil, nil // nolint:nilnil
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
	status, err := c.do(ctx, "GET", fmt.Sprintf("/autoscaling/groups/%s", id), nil, &group)
	if err != nil {
		return nil, err
	}

	if status == http.StatusNotFound {
		return nil, nil // nolint:nilnil
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

// Subnet represents the API response for a Subnet
type Subnet struct {
	ID               string `json:"id"`
	VPCID            string `json:"vpc_id"`
	Name             string `json:"name"`
	CIDRBlock        string `json:"cidr_block"`
	AvailabilityZone string `json:"availability_zone"`
}

func (c *Client) CreateSubnet(ctx context.Context, vpcID, name, cidr, az string) (*Subnet, error) {
	payload := map[string]string{
		"name":              name,
		"cidr_block":        cidr,
		"availability_zone": az,
	}

	var subnet Subnet
	_, err := c.do(ctx, "POST", fmt.Sprintf("/vpcs/%s/subnets", vpcID), payload, &subnet)
	if err != nil {
		return nil, err
	}

	return &subnet, nil
}

func (c *Client) GetSubnet(ctx context.Context, id string) (*Subnet, error) {
	var subnet Subnet
	status, err := c.do(ctx, "GET", fmt.Sprintf("/subnets/%s", id), nil, &subnet)
	if err != nil {
		return nil, err
	}

	if status == http.StatusNotFound {
		return nil, nil // nolint:nilnil
	}

	return &subnet, nil
}

func (c *Client) ListSubnets(ctx context.Context, vpcID string) ([]Subnet, error) {
	var subnets []Subnet
	_, err := c.do(ctx, "GET", fmt.Sprintf("/vpcs/%s/subnets", vpcID), nil, &subnets)
	if err != nil {
		return nil, err
	}
	return subnets, nil
}

func (c *Client) DeleteSubnet(ctx context.Context, id string) error {
	_, err := c.do(ctx, "DELETE", fmt.Sprintf("/subnets/%s", id), nil, nil)
	return err
}

// Snapshot represents the API response for a Snapshot
type Snapshot struct {
	ID          string `json:"id"`
	VolumeID    string `json:"volume_id"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

func (c *Client) CreateSnapshot(ctx context.Context, volumeID, description string) (*Snapshot, error) {
	payload := map[string]string{
		"volume_id":   volumeID,
		"description": description,
	}

	var snapshot Snapshot
	_, err := c.do(ctx, "POST", "/snapshots", payload, &snapshot)
	if err != nil {
		return nil, err
	}

	return &snapshot, nil
}

func (c *Client) GetSnapshot(ctx context.Context, id string) (*Snapshot, error) {
	var snapshot Snapshot
	status, err := c.do(ctx, "GET", fmt.Sprintf("/snapshots/%s", id), nil, &snapshot)
	if err != nil {
		return nil, err
	}

	if status == http.StatusNotFound {
		return nil, nil // nolint:nilnil
	}

	return &snapshot, nil
}

func (c *Client) ListSnapshots(ctx context.Context) ([]Snapshot, error) {
	var snapshots []Snapshot
	_, err := c.do(ctx, "GET", "/snapshots", nil, &snapshots)
	if err != nil {
		return nil, err
	}
	return snapshots, nil
}

func (c *Client) DeleteSnapshot(ctx context.Context, id string) error {
	_, err := c.do(ctx, "DELETE", fmt.Sprintf("/snapshots/%s", id), nil, nil)
	return err
}

// Database represents the API response for a Database
type Database struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Engine  string `json:"engine"`
	Version string `json:"version"`
	VpcID   string `json:"vpc_id,omitempty"`
	Status  string `json:"status"`
}

func (c *Client) CreateDatabase(ctx context.Context, name, engine, version, vpcID string) (*Database, error) {
	payload := map[string]interface{}{
		"name":    name,
		"engine":  engine,
		"version": version,
	}
	if vpcID != "" {
		payload["vpc_id"] = vpcID
	}

	var database Database
	_, err := c.do(ctx, "POST", "/databases", payload, &database)
	if err != nil {
		return nil, err
	}

	return &database, nil
}

func (c *Client) GetDatabase(ctx context.Context, id string) (*Database, error) {
	var database Database
	status, err := c.do(ctx, "GET", fmt.Sprintf("/databases/%s", id), nil, &database)
	if err != nil {
		return nil, err
	}

	if status == http.StatusNotFound {
		return nil, nil // nolint:nilnil
	}

	return &database, nil
}

func (c *Client) ListDatabases(ctx context.Context) ([]Database, error) {
	var databases []Database
	_, err := c.do(ctx, "GET", "/databases", nil, &databases)
	if err != nil {
		return nil, err
	}
	return databases, nil
}

func (c *Client) DeleteDatabase(ctx context.Context, id string) error {
	_, err := c.do(ctx, "DELETE", fmt.Sprintf("/databases/%s", id), nil, nil)
	return err
}

// ElasticIP represents the API response for an Elastic IP
type ElasticIP struct {
	ID         string `json:"id"`
	PublicIP   string `json:"public_ip"`
	InstanceID string `json:"instance_id,omitempty"`
	Status     string `json:"status"`
}

func (c *Client) AllocateElasticIP(ctx context.Context) (*ElasticIP, error) {
	var eip ElasticIP
	_, err := c.do(ctx, "POST", "/elastic-ips", nil, &eip)
	if err != nil {
		return nil, err
	}
	return &eip, nil
}

func (c *Client) GetElasticIP(ctx context.Context, id string) (*ElasticIP, error) {
	var eip ElasticIP
	status, err := c.do(ctx, "GET", fmt.Sprintf("/elastic-ips/%s", id), nil, &eip)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return nil, nil
	}
	return &eip, nil
}

func (c *Client) ListElasticIPs(ctx context.Context) ([]ElasticIP, error) {
	var eips []ElasticIP
	_, err := c.do(ctx, "GET", "/elastic-ips", nil, &eips)
	if err != nil {
		return nil, err
	}
	return eips, nil
}

func (c *Client) ReleaseElasticIP(ctx context.Context, id string) error {
	_, err := c.do(ctx, "DELETE", fmt.Sprintf("/elastic-ips/%s", id), nil, nil)
	return err
}

func (c *Client) AssociateElasticIP(ctx context.Context, id string, instanceID string) (*ElasticIP, error) {
	payload := map[string]string{
		"instance_id": instanceID,
	}
	var eip ElasticIP
	_, err := c.do(ctx, "POST", fmt.Sprintf("/elastic-ips/%s/associate", id), payload, &eip)
	if err != nil {
		return nil, err
	}
	return &eip, nil
}

func (c *Client) DisassociateElasticIP(ctx context.Context, id string) (*ElasticIP, error) {
	var eip ElasticIP
	_, err := c.do(ctx, "POST", fmt.Sprintf("/elastic-ips/%s/disassociate", id), nil, &eip)
	if err != nil {
		return nil, err
	}
	return &eip, nil
}

// DNSZone represents the API response for a DNS Zone
type DNSZone struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	VpcID       string `json:"vpc_id"`
	Status      string `json:"status"`
}

func (c *Client) CreateDNSZone(ctx context.Context, name, description, vpcID string) (*DNSZone, error) {
	payload := map[string]string{
		"name":        name,
		"description": description,
		"vpc_id":      vpcID,
	}
	var zone DNSZone
	_, err := c.do(ctx, "POST", "/dns/zones", payload, &zone)
	if err != nil {
		return nil, err
	}
	return &zone, nil
}

func (c *Client) GetDNSZone(ctx context.Context, id string) (*DNSZone, error) {
	var zone DNSZone
	status, err := c.do(ctx, "GET", fmt.Sprintf("/dns/zones/%s", id), nil, &zone)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return nil, nil
	}
	return &zone, nil
}

func (c *Client) ListDNSZones(ctx context.Context) ([]DNSZone, error) {
	var zones []DNSZone
	_, err := c.do(ctx, "GET", "/dns/zones", nil, &zones)
	if err != nil {
		return nil, err
	}
	return zones, nil
}

func (c *Client) DeleteDNSZone(ctx context.Context, id string) error {
	_, err := c.do(ctx, "DELETE", fmt.Sprintf("/dns/zones/%s", id), nil, nil)
	return err
}

// DNSRecord represents the API response for a DNS Record
type DNSRecord struct {
	ID       string `json:"id"`
	ZoneID   string `json:"zone_id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Content  string `json:"content"`
	TTL      int    `json:"ttl"`
	Priority *int   `json:"priority,omitempty"`
}

func (c *Client) CreateDNSRecord(ctx context.Context, zoneID string, record DNSRecord) (*DNSRecord, error) {
	var res DNSRecord
	_, err := c.do(ctx, "POST", fmt.Sprintf("/dns/zones/%s/records", zoneID), record, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Client) GetDNSRecord(ctx context.Context, id string) (*DNSRecord, error) {
	var record DNSRecord
	status, err := c.do(ctx, "GET", fmt.Sprintf("/dns/records/%s", id), nil, &record)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return nil, nil
	}
	return &record, nil
}

func (c *Client) ListDNSRecords(ctx context.Context, zoneID string) ([]DNSRecord, error) {
	var records []DNSRecord
	_, err := c.do(ctx, "GET", fmt.Sprintf("/dns/zones/%s/records", zoneID), nil, &records)
	if err != nil {
		return nil, err
	}
	return records, nil
}

func (c *Client) UpdateDNSRecord(ctx context.Context, id string, record DNSRecord) (*DNSRecord, error) {
	var res DNSRecord
	_, err := c.do(ctx, "PUT", fmt.Sprintf("/dns/records/%s", id), record, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Client) DeleteDNSRecord(ctx context.Context, id string) error {
	_, err := c.do(ctx, "DELETE", fmt.Sprintf("/dns/records/%s", id), nil, nil)
	return err
}

// Cluster represents the API response for a K8s Cluster
type Cluster struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	VpcID              string   `json:"vpc_id"`
	Version            string   `json:"version"`
	WorkerCount        int      `json:"worker_count"`
	Status             string   `json:"status"`
	PodCIDR            string   `json:"pod_cidr"`
	ServiceCIDR        string   `json:"service_cidr"`
	NetworkIsolation   bool     `json:"network_isolation"`
	HAEnabled          bool     `json:"ha_enabled"`
	APIServerLBAddress string   `json:"api_server_lb_address,omitempty"`
	ControlPlaneIPs    []string `json:"control_plane_ips"`
}

type CreateClusterRequest struct {
	Name             string `json:"name"`
	VpcID            string `json:"vpc_id"`
	Version          string `json:"version,omitempty"`
	Workers          int    `json:"workers,omitempty"`
	NetworkIsolation bool   `json:"network_isolation,omitempty"`
	HA               bool   `json:"ha,omitempty"`
}

func (c *Client) CreateCluster(ctx context.Context, req CreateClusterRequest) (*Cluster, error) {
	var cluster Cluster
	_, err := c.do(ctx, "POST", "/clusters", req, &cluster)
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (c *Client) GetCluster(ctx context.Context, id string) (*Cluster, error) {
	var cluster Cluster
	status, err := c.do(ctx, "GET", fmt.Sprintf("/clusters/%s", id), nil, &cluster)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return nil, nil
	}
	return &cluster, nil
}

func (c *Client) ListClusters(ctx context.Context) ([]Cluster, error) {
	var clusters []Cluster
	_, err := c.do(ctx, "GET", "/clusters", nil, &clusters)
	if err != nil {
		return nil, err
	}
	return clusters, nil
}

func (c *Client) DeleteCluster(ctx context.Context, id string) error {
	_, err := c.do(ctx, "DELETE", fmt.Sprintf("/clusters/%s", id), nil, nil)
	return err
}

func (c *Client) ScaleCluster(ctx context.Context, id string, workers int) error {
	payload := map[string]int{"workers": workers}
	_, err := c.do(ctx, "POST", fmt.Sprintf("/clusters/%s/scale", id), payload, nil)
	return err
}

func (c *Client) UpgradeCluster(ctx context.Context, id string, version string) error {
	payload := map[string]string{"version": version}
	_, err := c.do(ctx, "POST", fmt.Sprintf("/clusters/%s/upgrade", id), payload, nil)
	return err
}

// GlobalLB represents the API response for a Global Load Balancer
type GlobalLB struct {
	ID            string             `json:"id"`
	Name          string             `json:"name"`
	Hostname      string             `json:"hostname"`
	Policy        string             `json:"routing_policy"`
	Status        string             `json:"status"`
	HealthCheck   GlobalHealthCheck  `json:"health_check"`
	Endpoints     []GlobalEndpoint   `json:"endpoints,omitempty"`
}

type GlobalHealthCheck struct {
	Protocol       string `json:"protocol"`
	Port           int    `json:"port"`
	Path           string `json:"path,omitempty"`
	IntervalSec    int    `json:"interval_sec"`
	TimeoutSec     int    `json:"timeout_sec"`
	HealthyCount   int    `json:"healthy_count"`
	UnhealthyCount int    `json:"unhealthy_count"`
}

type GlobalEndpoint struct {
	ID         string `json:"id"`
	Region     string `json:"region"`
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id,omitempty"`
	TargetIP   string `json:"target_ip,omitempty"`
	Weight     int    `json:"weight"`
	Priority   int    `json:"priority"`
	Healthy    bool   `json:"healthy"`
}

type CreateGlobalLBRequest struct {
	Name        string            `json:"name"`
	Hostname    string            `json:"hostname"`
	Policy      string            `json:"policy"`
	HealthCheck GlobalHealthCheck `json:"health_check"`
}

func (c *Client) CreateGlobalLB(ctx context.Context, req CreateGlobalLBRequest) (*GlobalLB, error) {
	var glb GlobalLB
	_, err := c.do(ctx, "POST", "/global-lb", req, &glb)
	if err != nil {
		return nil, err
	}
	return &glb, nil
}

func (c *Client) GetGlobalLB(ctx context.Context, id string) (*GlobalLB, error) {
	var glb GlobalLB
	status, err := c.do(ctx, "GET", fmt.Sprintf("/global-lb/%s", id), nil, &glb)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return nil, nil
	}
	return &glb, nil
}

func (c *Client) ListGlobalLBs(ctx context.Context) ([]GlobalLB, error) {
	var glbs []GlobalLB
	_, err := c.do(ctx, "GET", "/global-lb", nil, &glbs)
	if err != nil {
		return nil, err
	}
	return glbs, nil
}

func (c *Client) DeleteGlobalLB(ctx context.Context, id string) error {
	_, err := c.do(ctx, "DELETE", fmt.Sprintf("/global-lb/%s", id), nil, nil)
	return err
}

type AddGlobalEndpointRequest struct {
	Region     string `json:"region"`
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id,omitempty"`
	TargetIP   string `json:"target_ip,omitempty"`
	Weight     int    `json:"weight,omitempty"`
	Priority   int    `json:"priority,omitempty"`
}

func (c *Client) AddGlobalEndpoint(ctx context.Context, glbID string, req AddGlobalEndpointRequest) (*GlobalEndpoint, error) {
	var ep GlobalEndpoint
	_, err := c.do(ctx, "POST", fmt.Sprintf("/global-lb/%s/endpoints", glbID), req, &ep)
	if err != nil {
		return nil, err
	}
	return &ep, nil
}

func (c *Client) RemoveGlobalEndpoint(ctx context.Context, glbID, epID string) error {
	_, err := c.do(ctx, "DELETE", fmt.Sprintf("/global-lb/%s/endpoints/%s", glbID, epID), nil, nil)
	return err
}
