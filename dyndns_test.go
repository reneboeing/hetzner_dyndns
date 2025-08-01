package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewDynDNSServer(t *testing.T) {
	client := NewClient("test-api-key")
	server := NewDynDNSServer(client, "admin", "password", "8080")

	if server.client != client {
		t.Error("Expected client to be set correctly")
	}
	if server.username != "admin" {
		t.Errorf("Expected username admin, got %s", server.username)
	}
	if server.password != "password" {
		t.Errorf("Expected password password, got %s", server.password)
	}
	if server.port != "8080" {
		t.Errorf("Expected port 8080, got %s", server.port)
	}
}

func TestHandleUpdateAuthentication(t *testing.T) {
	tests := []struct {
		name           string
		username       string
		password       string
		expectedStatus int
	}{
		{
			name:           "valid credentials",
			username:       "admin",
			password:       "password",
			expectedStatus: http.StatusBadRequest, // Will fail due to missing hostname, but auth passes
		},
		{
			name:           "invalid username",
			username:       "wrong",
			password:       "password",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid password",
			username:       "admin",
			password:       "wrong",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "no credentials",
			username:       "",
			password:       "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient("test-api-key")
			server := NewDynDNSServer(client, "admin", "password", "8080")

			req := httptest.NewRequest("GET", "/update", nil)
			if tt.username != "" || tt.password != "" {
				req.SetBasicAuth(tt.username, tt.password)
			}

			w := httptest.NewRecorder()
			server.handleUpdate(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusUnauthorized {
				if w.Header().Get("WWW-Authenticate") == "" {
					t.Error("Expected WWW-Authenticate header for 401 response")
				}
			}
		})
	}
}

func TestHandleUpdateMissingHostname(t *testing.T) {
	client := NewClient("test-api-key")
	server := NewDynDNSServer(client, "admin", "password", "8080")

	req := httptest.NewRequest("GET", "/update", nil)
	req.SetBasicAuth("admin", "password")

	w := httptest.NewRecorder()
	server.handleUpdate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Missing hostname") {
		t.Errorf("Expected error about missing hostname, got: %s", w.Body.String())
	}
}

func TestHandleUpdateOffline(t *testing.T) {
	client := NewClient("test-api-key")
	server := NewDynDNSServer(client, "admin", "password", "8080")

	req := httptest.NewRequest("GET", "/update?hostname=test.com&offline=yes", nil)
	req.SetBasicAuth("admin", "password")

	w := httptest.NewRecorder()
	server.handleUpdate(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "good" {
		t.Errorf("Expected response 'good', got '%s'", w.Body.String())
	}
}

func TestHandleUpdateIPValidation(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "invalid IPv4",
			query:          "hostname=test.com&myip=invalid.ip",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid IPv4 address",
		},
		{
			name:           "invalid IPv6",
			query:          "hostname=test.com&myipv6=invalid::ip::address",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid IPv6 address",
		},
		{
			name:           "no IP addresses",
			query:          "hostname=test.com",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "No valid IP address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient("test-api-key")
			server := NewDynDNSServer(client, "admin", "password", "8080")

			req := httptest.NewRequest("GET", "/update?"+tt.query, nil)
			req.SetBasicAuth("admin", "password")

			w := httptest.NewRecorder()
			server.handleUpdate(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("Expected body to contain '%s', got '%s'", tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestHandleUpdateWithMockClient(t *testing.T) {
	// Create a mock Hetzner API server
	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/zones":
			zones := ZonesResponse{
				Zones: []Zone{
					{ID: "zone123", Name: "example.com"},
				},
			}
			json.NewEncoder(w).Encode(zones)

		case r.URL.Path == "/records" && r.URL.Query().Get("zone_id") == "zone123":
			records := RecordsResponse{
				Records: []DNSRecord{
					{ID: "record123", Type: "A", Name: "test", Value: "1.2.3.4"},
				},
			}
			json.NewEncoder(w).Encode(records)

		case r.URL.Path == "/records/record123" && r.Method == "PUT":
			record := RecordResponse{
				Record: DNSRecord{ID: "record123", Type: "A", Name: "test", Value: "1.2.3.5"},
			}
			json.NewEncoder(w).Encode(record)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockAPI.Close()

	client := NewClient("test-api-key")
	client.BaseURL = mockAPI.URL

	server := NewDynDNSServer(client, "admin", "password", "8080")

	req := httptest.NewRequest("GET", "/update?hostname=test.example.com&myip=1.2.3.5", nil)
	req.SetBasicAuth("admin", "password")

	w := httptest.NewRecorder()
	server.handleUpdate(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	if !strings.Contains(w.Body.String(), "good") {
		t.Errorf("Expected success response, got '%s'", w.Body.String())
	}
}

func TestIsValidIPv4(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
	}{
		{"192.168.1.1", true},
		{"10.0.0.1", true},
		{"255.255.255.255", true},
		{"0.0.0.0", true},
		{"256.1.1.1", false},
		{"192.168.1", false},
		{"192.168.1.1.1", false},
		{"2001:db8::1", false},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			result := isValidIPv4(tt.ip)
			if result != tt.expected {
				t.Errorf("isValidIPv4(%s) = %v, expected %v", tt.ip, result, tt.expected)
			}
		})
	}
}

func TestIsValidIPv6(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
	}{
		{"2001:db8::1", true},
		{"2001:0db8:85a3:0000:0000:8a2e:0370:7334", true},
		{"::1", true},
		{"::ffff:192.0.2.1", true},
		{"192.168.1.1", false},
		{"invalid", false},
		{"2001:db8::g", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			result := isValidIPv6(tt.ip)
			if result != tt.expected {
				t.Errorf("isValidIPv6(%s) = %v, expected %v", tt.ip, result, tt.expected)
			}
		})
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expectedIP string
	}{
		{
			name:       "X-Forwarded-For header",
			headers:    map[string]string{"X-Forwarded-For": "192.168.1.1, 10.0.0.1"},
			remoteAddr: "127.0.0.1:12345",
			expectedIP: "192.168.1.1",
		},
		{
			name:       "X-Real-IP header",
			headers:    map[string]string{"X-Real-IP": "192.168.1.2"},
			remoteAddr: "127.0.0.1:12345",
			expectedIP: "192.168.1.2",
		},
		{
			name:       "RemoteAddr fallback",
			headers:    map[string]string{},
			remoteAddr: "192.168.1.3:12345",
			expectedIP: "192.168.1.3",
		},
		{
			name:       "RemoteAddr without port",
			headers:    map[string]string{},
			remoteAddr: "192.168.1.4",
			expectedIP: "192.168.1.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			req.RemoteAddr = tt.remoteAddr

			result := getClientIP(req)
			if result != tt.expectedIP {
				t.Errorf("Expected IP %s, got %s", tt.expectedIP, result)
			}
		})
	}
}

func TestUpdateDNSRecord(t *testing.T) {
	tests := []struct {
		name        string
		hostname    string
		ip          string
		recordType  string
		zones       []Zone
		records     []DNSRecord
		expectError bool
	}{
		{
			name:       "update existing record",
			hostname:   "test.example.com",
			ip:         "1.2.3.4",
			recordType: "A",
			zones: []Zone{
				{ID: "zone1", Name: "example.com"},
			},
			records: []DNSRecord{
				{ID: "rec1", Type: "A", Name: "test", Value: "1.2.3.3"},
			},
			expectError: false,
		},
		{
			name:       "create new record",
			hostname:   "new.example.com",
			ip:         "1.2.3.4",
			recordType: "A",
			zones: []Zone{
				{ID: "zone1", Name: "example.com"},
			},
			records:     []DNSRecord{},
			expectError: false,
		},
		{
			name:        "no matching zone",
			hostname:    "test.notfound.com",
			ip:          "1.2.3.4",
			recordType:  "A",
			zones:       []Zone{},
			records:     []DNSRecord{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.URL.Path == "/zones":
					response := ZonesResponse{Zones: tt.zones}
					json.NewEncoder(w).Encode(response)

				case strings.HasPrefix(r.URL.Path, "/records") && r.Method == "GET":
					response := RecordsResponse{Records: tt.records}
					json.NewEncoder(w).Encode(response)

				case strings.HasPrefix(r.URL.Path, "/records") && r.Method == "PUT":
					record := RecordResponse{
						Record: DNSRecord{ID: "updated", Type: tt.recordType, Name: "test", Value: tt.ip},
					}
					json.NewEncoder(w).Encode(record)

				case r.URL.Path == "/records" && r.Method == "POST":
					record := RecordResponse{
						Record: DNSRecord{ID: "created", Type: tt.recordType, Name: "test", Value: tt.ip},
					}
					json.NewEncoder(w).Encode(record)
				}
			}))
			defer mockAPI.Close()

			client := NewClient("test-api-key")
			client.BaseURL = mockAPI.URL

			server := NewDynDNSServer(client, "admin", "password", "8080")

			err := server.updateDNSRecord(tt.hostname, tt.ip, tt.recordType)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
