package main

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDNSRecordJSONMarshaling(t *testing.T) {
	ttl := 3600
	record := DNSRecord{
		ID:       "test123",
		Type:     "A",
		Name:     "test",
		Value:    "192.168.1.1",
		TTL:      &ttl,
		ZoneID:   "zone123",
		Created:  "2023-01-01T00:00:00Z",
		Modified: "2023-01-02T00:00:00Z",
	}

	// Test marshaling
	data, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("Failed to marshal DNSRecord: %v", err)
	}

	// Test unmarshaling
	var unmarshaled DNSRecord
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal DNSRecord: %v", err)
	}

	// Verify fields
	if unmarshaled.ID != record.ID {
		t.Errorf("Expected ID %s, got %s", record.ID, unmarshaled.ID)
	}
	if unmarshaled.Type != record.Type {
		t.Errorf("Expected Type %s, got %s", record.Type, unmarshaled.Type)
	}
	if unmarshaled.Name != record.Name {
		t.Errorf("Expected Name %s, got %s", record.Name, unmarshaled.Name)
	}
	if unmarshaled.Value != record.Value {
		t.Errorf("Expected Value %s, got %s", record.Value, unmarshaled.Value)
	}
	if *unmarshaled.TTL != *record.TTL {
		t.Errorf("Expected TTL %d, got %d", *record.TTL, *unmarshaled.TTL)
	}
}

func TestZoneJSONMarshaling(t *testing.T) {
	zone := Zone{
		ID:             "zone123",
		Name:           "example.com",
		TTL:            3600,
		Registrar:      "Example Registrar",
		LegacyDNSHost:  "old.dns.host",
		LegacyNS:       []string{"ns1.legacy.com", "ns2.legacy.com"},
		NS:             []string{"ns1.hetzner.com", "ns2.hetzner.com"},
		Created:        time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		Verified:       time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		Modified:       time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
		Project:        "project123",
		Owner:          "owner123",
		Permission:     "owner",
		ZoneType:       "native",
		Status:         "verified",
		Paused:         false,
		IsSecondaryDNS: false,
		RecordsCount:   5,
	}
	zone.TxtVerification.Name = "hetzner-domain-verification"
	zone.TxtVerification.Token = "verification-token-123"

	// Test marshaling
	data, err := json.Marshal(zone)
	if err != nil {
		t.Fatalf("Failed to marshal Zone: %v", err)
	}

	// Test unmarshaling
	var unmarshaled Zone
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal Zone: %v", err)
	}

	// Verify key fields
	if unmarshaled.ID != zone.ID {
		t.Errorf("Expected ID %s, got %s", zone.ID, unmarshaled.ID)
	}
	if unmarshaled.Name != zone.Name {
		t.Errorf("Expected Name %s, got %s", zone.Name, unmarshaled.Name)
	}
	if unmarshaled.TTL != zone.TTL {
		t.Errorf("Expected TTL %d, got %d", zone.TTL, unmarshaled.TTL)
	}
	if len(unmarshaled.NS) != len(zone.NS) {
		t.Errorf("Expected %d NS records, got %d", len(zone.NS), len(unmarshaled.NS))
	}
}

func TestCreateRecordRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		request CreateRecordRequest
		valid   bool
	}{
		{
			name: "valid A record",
			request: CreateRecordRequest{
				Type:   "A",
				Name:   "test",
				Value:  "192.168.1.1",
				ZoneID: "zone123",
			},
			valid: true,
		},
		{
			name: "valid AAAA record",
			request: CreateRecordRequest{
				Type:   "AAAA",
				Name:   "test",
				Value:  "2001:db8::1",
				ZoneID: "zone123",
			},
			valid: true,
		},
		{
			name: "valid CNAME record",
			request: CreateRecordRequest{
				Type:   "CNAME",
				Name:   "www",
				Value:  "example.com",
				ZoneID: "zone123",
			},
			valid: true,
		},
		{
			name: "valid MX record",
			request: CreateRecordRequest{
				Type:   "MX",
				Name:   "@",
				Value:  "10 mail.example.com",
				ZoneID: "zone123",
			},
			valid: true,
		},
		{
			name: "valid TXT record",
			request: CreateRecordRequest{
				Type:   "TXT",
				Name:   "_dmarc",
				Value:  "v=DMARC1; p=none",
				ZoneID: "zone123",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that the request can be marshaled/unmarshaled
			data, err := json.Marshal(tt.request)
			if err != nil {
				if tt.valid {
					t.Errorf("Failed to marshal valid request: %v", err)
				}
				return
			}

			var unmarshaled CreateRecordRequest
			err = json.Unmarshal(data, &unmarshaled)
			if err != nil {
				if tt.valid {
					t.Errorf("Failed to unmarshal valid request: %v", err)
				}
				return
			}

			// Verify required fields are preserved
			if unmarshaled.Type != tt.request.Type {
				t.Errorf("Expected Type %s, got %s", tt.request.Type, unmarshaled.Type)
			}
			if unmarshaled.Name != tt.request.Name {
				t.Errorf("Expected Name %s, got %s", tt.request.Name, unmarshaled.Name)
			}
			if unmarshaled.Value != tt.request.Value {
				t.Errorf("Expected Value %s, got %s", tt.request.Value, unmarshaled.Value)
			}
			if unmarshaled.ZoneID != tt.request.ZoneID {
				t.Errorf("Expected ZoneID %s, got %s", tt.request.ZoneID, unmarshaled.ZoneID)
			}
		})
	}
}

func TestUpdateRecordRequestValidation(t *testing.T) {
	ttl := 7200
	request := UpdateRecordRequest{
		Type:  "A",
		Name:  "updated",
		Value: "192.168.1.100",
		TTL:   &ttl,
	}

	// Test marshaling
	data, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal UpdateRecordRequest: %v", err)
	}

	// Test unmarshaling
	var unmarshaled UpdateRecordRequest
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal UpdateRecordRequest: %v", err)
	}

	// Verify fields
	if unmarshaled.Type != request.Type {
		t.Errorf("Expected Type %s, got %s", request.Type, unmarshaled.Type)
	}
	if unmarshaled.Name != request.Name {
		t.Errorf("Expected Name %s, got %s", request.Name, unmarshaled.Name)
	}
	if unmarshaled.Value != request.Value {
		t.Errorf("Expected Value %s, got %s", request.Value, unmarshaled.Value)
	}
	if *unmarshaled.TTL != *request.TTL {
		t.Errorf("Expected TTL %d, got %d", *request.TTL, *unmarshaled.TTL)
	}
}

func TestRecordsResponseStructure(t *testing.T) {
	jsonData := `{
		"records": [
			{
				"id": "rec1",
				"type": "A",
				"name": "test1",
				"value": "192.168.1.1",
				"zone_id": "zone123"
			},
			{
				"id": "rec2",
				"type": "AAAA",
				"name": "test2",
				"value": "2001:db8::1",
				"zone_id": "zone123"
			}
		]
	}`

	var response RecordsResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal RecordsResponse: %v", err)
	}

	if len(response.Records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(response.Records))
	}

	if response.Records[0].ID != "rec1" {
		t.Errorf("Expected first record ID rec1, got %s", response.Records[0].ID)
	}
}

func TestAPIErrorStructure(t *testing.T) {
	jsonData := `{
		"error": {
			"message": "Invalid request parameters",
			"code": 400
		}
	}`

	var apiError APIError
	err := json.Unmarshal([]byte(jsonData), &apiError)
	if err != nil {
		t.Fatalf("Failed to unmarshal APIError: %v", err)
	}

	if apiError.Error.Message != "Invalid request parameters" {
		t.Errorf("Expected message 'Invalid request parameters', got '%s'", apiError.Error.Message)
	}

	if apiError.Error.Code != 400 {
		t.Errorf("Expected code 400, got %d", apiError.Error.Code)
	}
}

func TestRecordResponseStructure(t *testing.T) {
	jsonData := `{
		"record": {
			"id": "rec123",
			"type": "A",
			"name": "test",
			"value": "192.168.1.1",
			"ttl": 3600,
			"zone_id": "zone123",
			"created": "2023-01-01T00:00:00Z",
			"modified": "2023-01-02T00:00:00Z"
		}
	}`

	var response RecordResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal RecordResponse: %v", err)
	}

	if response.Record.ID != "rec123" {
		t.Errorf("Expected record ID rec123, got %s", response.Record.ID)
	}

	if response.Record.Type != "A" {
		t.Errorf("Expected record type A, got %s", response.Record.Type)
	}

	if response.Record.TTL == nil || *response.Record.TTL != 3600 {
		t.Errorf("Expected TTL 3600, got %v", response.Record.TTL)
	}
}

func TestZoneResponseStructure(t *testing.T) {
	jsonData := `{
		"zones": [
			{
				"id": "zone123",
				"name": "example.com",
				"ttl": 3600,
				"registrar": "Example Registrar",
				"legacy_dns_host": "old.dns.host",
				"legacy_ns": ["ns1.legacy.com", "ns2.legacy.com"],
				"ns": ["ns1.hetzner.com", "ns2.hetzner.com"],
				"created": "2023-01-01T00:00:00Z",
				"verified": "2023-01-01T12:00:00Z",
				"modified": "2023-01-02T00:00:00Z",
				"project": "project123",
				"owner": "owner123",
				"permission": "owner",
				"zone_type": "native",
				"status": "verified",
				"paused": false,
				"is_secondary_dns": false,
				"txt_verification": {
					"name": "hetzner-domain-verification",
					"token": "verification-token-123"
				},
				"records_count": 5
			}
		],
		"meta": {
			"pagination": {
				"page": 1,
				"per_page": 25,
				"previous_page": 0,
				"next_page": 0,
				"last_page": 1,
				"total_entries": 1
			}
		}
	}`

	var response ZonesResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal ZonesResponse: %v", err)
	}

	if len(response.Zones) != 1 {
		t.Errorf("Expected 1 zone, got %d", len(response.Zones))
	}

	zone := response.Zones[0]
	if zone.ID != "zone123" {
		t.Errorf("Expected zone ID zone123, got %s", zone.ID)
	}

	if zone.Name != "example.com" {
		t.Errorf("Expected zone name example.com, got %s", zone.Name)
	}

	if len(zone.NS) != 2 {
		t.Errorf("Expected 2 nameservers, got %d", len(zone.NS))
	}

	if zone.TxtVerification.Token != "verification-token-123" {
		t.Errorf("Expected verification token verification-token-123, got %s", zone.TxtVerification.Token)
	}
}
