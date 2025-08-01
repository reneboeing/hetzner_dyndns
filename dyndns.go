package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// DynDNSServer handles DynDNS update requests from FritzBox
type DynDNSServer struct {
	client   *Client
	username string
	password string
	port     string
}

// NewDynDNSServer creates a new DynDNS server
func NewDynDNSServer(client *Client, username, password, port string) *DynDNSServer {
	return &DynDNSServer{
		client:   client,
		username: username,
		password: password,
		port:     port,
	}
}

// handleUpdate handles DynDNS update requests
func (s *DynDNSServer) handleUpdate(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	user, pass, ok := r.BasicAuth()
	if !ok || user != s.username || pass != s.password {
		w.Header().Set("WWW-Authenticate", `Basic realm="DynDNS"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	hostname := r.URL.Query().Get("hostname")
	myip := r.URL.Query().Get("myip")
	myipv6 := r.URL.Query().Get("myipv6")
	offline := r.URL.Query().Get("offline")

	log.Printf("DynDNS update request: hostname=%s, myip=%s, myipv6=%s, offline=%s", hostname, myip, myipv6, offline)

	if hostname == "" {
		http.Error(w, "Missing hostname parameter", http.StatusBadRequest)
		return
	}

	// Handle offline request
	if offline == "yes" {
		log.Printf("Offline request for %s - not implemented", hostname)
		fmt.Fprintf(w, "good")
		return
	}

	var ipv4, ipv6 string
	var updateResults []string

	// Handle IPv4 address
	if myip != "" {
		if isValidIPv4(myip) {
			ipv4 = myip
		} else {
			http.Error(w, "Invalid IPv4 address", http.StatusBadRequest)
			return
		}
	}

	// Handle IPv6 address
	if myipv6 != "" {
		if isValidIPv6(myipv6) {
			ipv6 = myipv6
		} else {
			http.Error(w, "Invalid IPv6 address", http.StatusBadRequest)
			return
		}
	}

	// If no IP addresses provided and we couldn't detect any, error
	if ipv4 == "" && ipv6 == "" {
		http.Error(w, "No valid IP address provided or detected", http.StatusBadRequest)
		return
	}

	// Update IPv4 record if provided
	if ipv4 != "" {
		err := s.updateDNSRecord(hostname, ipv4, "A")
		if err != nil {
			log.Printf("Failed to update IPv4 DNS record: %v", err)
			fmt.Fprintf(w, "911")
			return
		}
		updateResults = append(updateResults, fmt.Sprintf("IPv4: %s", ipv4))
		log.Printf("Successfully updated %s A record to %s", hostname, ipv4)
	}

	// Update IPv6 record if provided
	if ipv6 != "" {
		err := s.updateDNSRecord(hostname, ipv6, "AAAA")
		if err != nil {
			log.Printf("Failed to update IPv6 DNS record: %v", err)
			fmt.Fprintf(w, "911")
			return
		}
		updateResults = append(updateResults, fmt.Sprintf("IPv6: %s", ipv6))
		log.Printf("Successfully updated %s AAAA record to %s", hostname, ipv6)
	}

	// Return success response with the updated IPs
	if len(updateResults) > 0 {
		fmt.Fprintf(w, "good %s", strings.Join(updateResults, ", "))
	} else {
		fmt.Fprintf(w, "good")
	}
}

// handleHealth handles health check requests
func (s *DynDNSServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	// Simple health check - verify the server is responding
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status":    "healthy",
		"service":   "hetzner-dns-bridge",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "1.0.0",
	}

	json.NewEncoder(w).Encode(response)
}

// updateDNSRecord updates the DNS record using Hetzner API
func (s *DynDNSServer) updateDNSRecord(hostname, ip, recordType string) error {
	// Get all zones to find the correct one
	zones, err := s.client.GetZones()
	if err != nil {
		return fmt.Errorf("failed to get zones: %w", err)
	}

	var targetZone *Zone
	var recordName string

	// Find the zone that matches the hostname
	for _, zone := range zones {
		if hostname == zone.Name {
			// Exact match - update root record
			targetZone = &zone
			recordName = "@"
			break
		} else if strings.HasSuffix(hostname, "."+zone.Name) {
			// Subdomain - extract the subdomain part
			targetZone = &zone
			recordName = strings.TrimSuffix(hostname, "."+zone.Name)
			break
		}
	}

	if targetZone == nil {
		return fmt.Errorf("no zone found for hostname: %s", hostname)
	}

	log.Printf("Found zone: %s (ID: %s) for hostname: %s, record name: %s",
		targetZone.Name, targetZone.ID, hostname, recordName)

	// Get existing records for the zone
	records, err := s.client.GetAllRecords(targetZone.ID)
	if err != nil {
		return fmt.Errorf("failed to get records: %w", err)
	}

	// Look for existing record
	var existingRecord *DNSRecord
	for _, record := range records {
		if record.Name == recordName && record.Type == recordType {
			existingRecord = &record
			break
		}
	}

	if existingRecord != nil {
		// Update existing record
		updateReq := UpdateRecordRequest{
			ZoneID: targetZone.ID,
			Type:  recordType,
			Name:  recordName,
			Value: ip,
			TTL:   existingRecord.TTL,
		}

		log.Printf("UpdateRecord %v",
			   updateReq)

		_, err = s.client.UpdateRecord(existingRecord.ID, updateReq)
		if err != nil {
			return fmt.Errorf("failed to update record: %w", err)
		}

		log.Printf("Updated existing record %s (%s) to %s", existingRecord.ID, recordType, ip)
	} else {
		// Create new record
		ttl := 3600 // 5 minutes TTL for dynamic records
		createReq := CreateRecordRequest{
			Type:   recordType,
			Name:   recordName,
			Value:  ip,
			TTL:    ttl,
			ZoneID: targetZone.ID,
		}


		log.Printf("createReq %v",
			   createReq)
		_, err = s.client.CreateRecord(createReq)
		if err != nil {
			return fmt.Errorf("failed to create record: %w", err)
		}

		log.Printf("Created new record %s %s -> %s", recordType, recordName, ip)
	}

	return nil
}

// isValidIPv4 checks if the given string is a valid IPv4 address
func isValidIPv4(ip string) bool {
	return net.ParseIP(ip) != nil && strings.Count(ip, ":") == 0
}

// isValidIPv6 checks if the given string is a valid IPv6 address
func isValidIPv6(ip string) bool {
	return net.ParseIP(ip) != nil && strings.Count(ip, ":") > 0
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP if there are multiple
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// Start starts the DynDNS server
func (s *DynDNSServer) Start() error {
	http.HandleFunc("/update", s.handleUpdate)
	http.HandleFunc("/nic/update", s.handleUpdate) // Alternative endpoint some clients use
	http.HandleFunc("/health", s.handleHealth)     // Health check endpoint
	http.HandleFunc("/", s.handleHealth)           // Root endpoint for simple health checks

	log.Printf("Starting DynDNS server on port %s", s.port)
	log.Printf("Update URL: http://localhost:%s/update?hostname=yourdomain.com&myip=1.2.3.4", s.port)
	log.Printf("Configure your FritzBox with:")
	log.Printf("  Update URL: http://your-server:%s/update", s.port)
	log.Printf("  Domain: <your-domain>")
	log.Printf("  Username: %s", s.username)
	log.Printf("  Password: %s", s.password)

	return http.ListenAndServe(":"+s.port, nil)
}
