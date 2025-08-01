package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"io"
	"net/http"
	"time"
)

const (
	BaseURL = "https://dns.hetzner.com/api/v1"
)

// Client represents the Hetzner DNS API client
type Client struct {
	APIKey     string
	HTTPClient *http.Client
	BaseURL    string
}

// NewClient creates a new Hetzner DNS API client
func NewClient(apiKey string) *Client {
	return &Client{
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		BaseURL: BaseURL,
	}
}

// makeRequest makes an HTTP request to the Hetzner DNS API
func (c *Client) makeRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	log.Printf("Execute request to '%s' using body '%s'", endpoint, reqBody)
	req, err := http.NewRequest(method, c.BaseURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Auth-API-Token", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	return resp, nil
}

// handleResponse handles the HTTP response and unmarshals JSON
func (c *Client) handleResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiError APIError
		if err := json.Unmarshal(body, &apiError); err != nil {
			return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
		}

		log.Printf("Error from API: '%d' using body '%s'", resp.StatusCode, string(body))
		return fmt.Errorf("API error: %s", apiError.Error.Message)
	}

	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// GetAllRecords retrieves all DNS records for a zone
func (c *Client) GetAllRecords(zoneID string) ([]DNSRecord, error) {
	endpoint := fmt.Sprintf("/records?zone_id=%s", zoneID)

	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var recordsResp RecordsResponse
	if err := c.handleResponse(resp, &recordsResp); err != nil {
		return nil, err
	}

	return recordsResp.Records, nil
}

// GetRecord retrieves a specific DNS record by ID
func (c *Client) GetRecord(recordID string) (*DNSRecord, error) {
	endpoint := fmt.Sprintf("/records/%s", recordID)

	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var recordResp RecordResponse
	if err := c.handleResponse(resp, &recordResp); err != nil {
		return nil, err
	}

	return &recordResp.Record, nil
}

// CreateRecord creates a new DNS record
func (c *Client) CreateRecord(req CreateRecordRequest) (*DNSRecord, error) {
	resp, err := c.makeRequest("POST", "/records", req)
	if err != nil {
		return nil, err
	}

	var recordResp RecordResponse
	if err := c.handleResponse(resp, &recordResp); err != nil {
		return nil, err
	}

	return &recordResp.Record, nil
}

// UpdateRecord updates an existing DNS record
func (c *Client) UpdateRecord(recordID string, req UpdateRecordRequest) (*DNSRecord, error) {
	endpoint := fmt.Sprintf("/records/%s", recordID)

	resp, err := c.makeRequest("PUT", endpoint, req)
	if err != nil {
		return nil, err
	}

	var recordResp RecordResponse
	if err := c.handleResponse(resp, &recordResp); err != nil {
		return nil, err
	}

	return &recordResp.Record, nil
}

// DeleteRecord deletes a DNS record by ID
func (c *Client) DeleteRecord(recordID string) error {
	endpoint := fmt.Sprintf("/records/%s", recordID)

	resp, err := c.makeRequest("DELETE", endpoint, nil)
	if err != nil {
		return err
	}

	return c.handleResponse(resp, nil)
}

// GetZones retrieves all DNS zones
func (c *Client) GetZones() ([]Zone, error) {
	resp, err := c.makeRequest("GET", "/zones", nil)
	if err != nil {
		return nil, err
	}

	var zonesResp ZonesResponse
	if err := c.handleResponse(resp, &zonesResp); err != nil {
		return nil, err
	}

	return zonesResp.Zones, nil
}
