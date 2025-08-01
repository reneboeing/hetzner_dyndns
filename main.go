package main

import (
	"log"
	"os"
)

func main() {
	// Get API key from environment variable
	apiKey := os.Getenv("HETZNER_DNS_API_KEY")
	if apiKey == "" {
		log.Fatal("HETZNER_DNS_API_KEY environment variable is required")
	}

	// Get DynDNS credentials from environment
	username := os.Getenv("DYNDNS_USERNAME")
	if username == "" {
		username = "admin" // Default username
	}

	password := os.Getenv("DYNDNS_PASSWORD")
	if password == "" {
		log.Fatal("DYNDNS_PASSWORD environment variable is required")
	}

	port := os.Getenv("DYNDNS_PORT")
	if port == "" {
		port = "8080" // Default port
	}

	// Create Hetzner DNS client
	client := NewClient(apiKey)

	// Create and start DynDNS server
	server := NewDynDNSServer(client, username, password, port)

	log.Printf("Starting DynDNS bridge for FritzBox -> Hetzner DNS")
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
