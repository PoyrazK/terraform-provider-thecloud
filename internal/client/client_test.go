package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testKey     = "test-key"
	testVpcID   = "vpc-123"
	testVpcName = "test-vpc"
	testCIDR    = "10.0.0.0/16"
)

func TestClientCreateVPC(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/vpcs", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, testKey, r.Header.Get("X-API-Key"))

		var payload map[string]string
		json.NewDecoder(r.Body).Decode(&payload)
		assert.Equal(t, testVpcName, payload["name"])

		w.WriteHeader(http.StatusCreated)
		data, _ := json.Marshal(VPC{
			ID:        testVpcID,
			Name:      testVpcName,
			CIDRBlock: testCIDR,
			Status:    "available",
		})
		json.NewEncoder(w).Encode(APIResponse{
			Data: data,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, testKey)
	vpc, err := c.CreateVPC(context.Background(), testVpcName, testCIDR)

	assert.NoError(t, err)
	assert.NotNil(t, vpc)
	assert.Equal(t, testVpcID, vpc.ID)
	assert.Equal(t, testVpcName, vpc.Name)
}

func TestClientGetVPC(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/vpcs/"+testVpcID, r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.WriteHeader(http.StatusOK)
		data, _ := json.Marshal(VPC{
			ID:        testVpcID,
			Name:      testVpcName,
			CIDRBlock: testCIDR,
			Status:    "available",
		})
		json.NewEncoder(w).Encode(APIResponse{
			Data: data,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, testKey)
	vpc, err := c.GetVPC(context.Background(), testVpcID)

	assert.NoError(t, err)
	assert.NotNil(t, vpc)
	assert.Equal(t, testVpcID, vpc.ID)
}

func TestClientGetVPCNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c := NewClient(server.URL, testKey)
	vpc, err := c.GetVPC(context.Background(), "non-existent")

	assert.NoError(t, err)
	assert.Nil(t, vpc)
}

func TestClientDeleteVPC(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/vpcs/"+testVpcID, r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(APIResponse{
			Data: json.RawMessage(`{"message": "deleted"}`),
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, testKey)
	err := c.DeleteVPC(context.Background(), testVpcID)

	assert.NoError(t, err)
}

func TestClientError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{
			Error: &APIError{
				Type:    "invalid_input",
				Message: "invalid cidr",
				Code:    "400",
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, testKey)
	_, err := c.CreateVPC(context.Background(), testVpcName, "invalid")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid cidr")
}
