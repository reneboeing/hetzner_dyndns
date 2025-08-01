package main

import "time"

// DNSRecord represents a DNS record in the Hetzner DNS API
type DNSRecord struct {
	ID       string `json:"id,omitempty"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	TTL      *int   `json:"ttl,omitempty"`
	ZoneID   string `json:"zone_id,omitempty"`
	Created  string `json:"created,omitempty"`
	Modified string `json:"modified,omitempty"`
}

// Zone represents a DNS zone
type Zone struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	TTL             int       `json:"ttl"`
	Registrar       string    `json:"registrar"`
	LegacyDNSHost   string    `json:"legacy_dns_host"`
	LegacyNS        []string  `json:"legacy_ns"`
	NS              []string  `json:"ns"`
	Created         time.Time `json:"created"`
	Verified        time.Time `json:"verified"`
	Modified        time.Time `json:"modified"`
	Project         string    `json:"project"`
	Owner           string    `json:"owner"`
	Permission      string    `json:"permission"`
	ZoneType        string    `json:"zone_type"`
	Status          string    `json:"status"`
	Paused          bool      `json:"paused"`
	IsSecondaryDNS  bool      `json:"is_secondary_dns"`
	TxtVerification struct {
		Name  string `json:"name"`
		Token string `json:"token"`
	} `json:"txt_verification"`
	RecordsCount int `json:"records_count"`
}

// RecordsResponse represents the response when getting multiple records
type RecordsResponse struct {
	Records []DNSRecord `json:"records"`
}

// RecordResponse represents the response when getting/creating/updating a single record
type RecordResponse struct {
	Record DNSRecord `json:"record"`
}

// ZonesResponse represents the response when getting zones
type ZonesResponse struct {
	Zones []Zone `json:"zones"`
	Meta  struct {
		Pagination struct {
			Page         int `json:"page"`
			PerPage      int `json:"per_page"`
			PreviousPage int `json:"previous_page"`
			NextPage     int `json:"next_page"`
			LastPage     int `json:"last_page"`
			TotalEntries int `json:"total_entries"`
		} `json:"pagination"`
	} `json:"meta"`
}

// APIError represents an error response from the API
type APIError struct {
	Error struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error"`
}

// CreateRecordRequest represents the request to create a new record
type CreateRecordRequest struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Value  string `json:"value"`
	TTL    *int   `json:"ttl,omitempty"`
	ZoneID string `json:"zone_id"`
}

// UpdateRecordRequest represents the request to update a record
type UpdateRecordRequest struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
	TTL   *int   `json:"ttl,omitempty"`
}
