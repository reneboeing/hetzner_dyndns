package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	apiKey := "test-api-key"
	client := NewClient(apiKey)

	if client.APIKey != apiKey {
		t.Errorf("Expected API key %s, got %s", apiKey, client.APIKey)
	}

	if client.BaseURL != BaseURL {
		t.Errorf("Expected base URL %s, got %s", BaseURL, client.BaseURL)
	}

	if client.HTTPClient.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", client.HTTPClient.Timeout)
	}
}

func TestMakeRequest(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		endpoint       string
		body           interface{}
		expectedMethod string
		expectedHeader string
	}{
		{
			name:           "GET request",
			method:         "GET",
			endpoint:       "/zones",
			body:           nil,
			expectedMethod: "GET",
			expectedHeader: "test-api-key",
		},
		{
			name:     "POST request with body",
			method:   "POST",
			endpoint: "/records",
			body: CreateRecordRequest{
				Type:   "A",
				Name:   "test",
				Value:  "1.2.3.4",
				ZoneID: "zone123",
			},
			expectedMethod: "POST",
			expectedHeader: "test-api-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != tt.expectedMethod {
					t.Errorf("Expected method %s, got %s", tt.expectedMethod, r.Method)
				}

				if r.Header.Get("Auth-API-Token") != tt.expectedHeader {
					t.Errorf("Expected Auth-API-Token %s, got %s", tt.expectedHeader, r.Header.Get("Auth-API-Token"))
				}

				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
				}

				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client := NewClient("test-api-key")
			client.BaseURL = server.URL

			resp, err := client.makeRequest(tt.method, tt.endpoint, tt.body)
			if err != nil {
				t.Fatalf("makeRequest failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200, got %d", resp.StatusCode)
			}
		})
	}
}

func TestHandleResponse(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		responseBody  string
		expectError   bool
		errorContains string
	}{
		{
			name:         "success response",
			statusCode:   200,
			responseBody: `{"record":{"id":"123","type":"A","name":"test","value":"1.2.3.4"}}`,
			expectError:  false,
		},
		{
			name:          "API error response",
			statusCode:    400,
			responseBody:  `{"error":{"message":"Invalid request","code":400}}`,
			expectError:   true,
			errorContains: "Invalid request",
		},
		{
			name:          "non-JSON error response",
			statusCode:    500,
			responseBody:  "Internal Server Error",
			expectError:   true,
			errorContains: "API request failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewClient("test-api-key")
			resp, err := http.Get(server.URL)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}

			var result RecordResponse
			err = client.handleResponse(resp, &result)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestGetAllRecords(t *testing.T) {
	mockRecords := RecordsResponse{
		Records: []DNSRecord{
			{ID: "1", Type: "A", Name: "test", Value: "1.2.3.4"},
			{ID: "2", Type: "AAAA", Name: "test", Value: "2001:db8::1"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/records" {
			t.Errorf("Expected path /records, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("zone_id") != "zone123" {
			t.Errorf("Expected zone_id zone123, got %s", r.URL.Query().Get("zone_id"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockRecords)
	}))
	defer server.Close()

	client := NewClient("test-api-key")
	client.BaseURL = server.URL

	records, err := client.GetAllRecords("zone123")
	if err != nil {
		t.Fatalf("GetAllRecords failed: %v", err)
	}

	if len(records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(records))
	}

	if records[0].ID != "1" || records[0].Type != "A" {
		t.Errorf("Unexpected first record: %+v", records[0])
	}
}

func TestGetRecord(t *testing.T) {
	mockRecord := RecordResponse{
		Record: DNSRecord{ID: "123", Type: "A", Name: "test", Value: "1.2.3.4"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/records/123" {
			t.Errorf("Expected path /records/123, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockRecord)
	}))
	defer server.Close()

	client := NewClient("test-api-key")
	client.BaseURL = server.URL

	record, err := client.GetRecord("123")
	if err != nil {
		t.Fatalf("GetRecord failed: %v", err)
	}

	if record.ID != "123" || record.Type != "A" {
		t.Errorf("Unexpected record: %+v", record)
	}
}

func TestCreateRecord(t *testing.T) {
	ttl := 3600
	createReq := CreateRecordRequest{
		Type:   "A",
		Name:   "test",
		Value:  "1.2.3.4",
		TTL:    &ttl,
		ZoneID: "zone123",
	}

	mockResponse := RecordResponse{
		Record: DNSRecord{ID: "new123", Type: "A", Name: "test", Value: "1.2.3.4"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/records" {
			t.Errorf("Expected POST /records, got %s %s", r.Method, r.URL.Path)
		}

		var receivedReq CreateRecordRequest
		if err := json.NewDecoder(r.Body).Decode(&receivedReq); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if receivedReq.Type != createReq.Type || receivedReq.Name != createReq.Name {
			t.Errorf("Unexpected request: %+v", receivedReq)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewClient("test-api-key")
	client.BaseURL = server.URL

	record, err := client.CreateRecord(createReq)
	if err != nil {
		t.Fatalf("CreateRecord failed: %v", err)
	}

	if record.ID != "new123" || record.Type != "A" {
		t.Errorf("Unexpected record: %+v", record)
	}
}

func TestUpdateRecord(t *testing.T) {
	ttl := 3600
	updateReq := UpdateRecordRequest{
		Type:  "A",
		Name:  "test",
		Value: "1.2.3.5",
		TTL:   &ttl,
	}

	mockResponse := RecordResponse{
		Record: DNSRecord{ID: "123", Type: "A", Name: "test", Value: "1.2.3.5"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" || r.URL.Path != "/records/123" {
			t.Errorf("Expected PUT /records/123, got %s %s", r.Method, r.URL.Path)
		}

		var receivedReq UpdateRecordRequest
		if err := json.NewDecoder(r.Body).Decode(&receivedReq); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if receivedReq.Value != updateReq.Value {
			t.Errorf("Expected value %s, got %s", updateReq.Value, receivedReq.Value)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewClient("test-api-key")
	client.BaseURL = server.URL

	record, err := client.UpdateRecord("123", updateReq)
	if err != nil {
		t.Fatalf("UpdateRecord failed: %v", err)
	}

	if record.Value != "1.2.3.5" {
		t.Errorf("Expected value 1.2.3.5, got %s", record.Value)
	}
}

func TestDeleteRecord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/records/123" {
			t.Errorf("Expected DELETE /records/123, got %s %s", r.Method, r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("test-api-key")
	client.BaseURL = server.URL

	err := client.DeleteRecord("123")
	if err != nil {
		t.Fatalf("DeleteRecord failed: %v", err)
	}
}

func TestGetZones(t *testing.T) {
	mockZones := ZonesResponse{
		Zones: []Zone{
			{ID: "zone1", Name: "example.com", TTL: 3600},
			{ID: "zone2", Name: "test.com", TTL: 7200},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/zones" {
			t.Errorf("Expected path /zones, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockZones)
	}))
	defer server.Close()

	client := NewClient("test-api-key")
	client.BaseURL = server.URL

	zones, err := client.GetZones()
	if err != nil {
		t.Fatalf("GetZones failed: %v", err)
	}

	if len(zones) != 2 {
		t.Errorf("Expected 2 zones, got %d", len(zones))
	}

	if zones[0].Name != "example.com" || zones[1].Name != "test.com" {
		t.Errorf("Unexpected zones: %+v", zones)
	}
}
