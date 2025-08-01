# Hetzner DNS API Client with FritzBox DynDNS Bridge

A Go implementation of a Hetzner DNS API client with a built-in DynDNS server that accepts FritzBox-compatible update requests and manages DNS records through the Hetzner DNS API.

## Features

### Hetzner DNS API Client
- ✅ Full CRUD operations for DNS records (Create, Read, Update, Delete)
- ✅ API key authentication
- ✅ Support for all DNS record types (A, AAAA, CNAME, MX, TXT, etc.)

### FritzBox DynDNS Bridge
- ✅ FritzBox-compatible DynDNS update endpoint
- ✅ HTTP Basic Authentication
- ✅ IPv4 and IPv6 support (dual-stack)
- ✅ Automatic zone detection
- ✅ Subdomain support
- ✅ Update existing records or create new ones

## Quick Start

### Prerequisites
- Go 1.24 or later
- Hetzner DNS API token
- Domain hosted on Hetzner DNS

### Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/hetzner-dns-api.git
cd hetzner-dns-api
```

2. Build the application:
```bash
go build -o hetzner-dns-bridge
```

### Configuration

Set the required environment variables:

```bash
# Required
export HETZNER_DNS_API_KEY="your_hetzner_api_token_here"
export DYNDNS_PASSWORD="your_secure_password"

# Optional (with defaults)
export DYNDNS_USERNAME="admin"  # Default: admin
export DYNDNS_PORT="8080"       # Default: 8080
```
The password and username are used for FritzBox authentication. You may choose any non empty combination, but it is recommended to use a strong password.
The port is where the DynDNS server will listen for requests.
### Running the Server

```bash
./hetzner-dns-bridge
```

The server will start and display configuration information:
```
Starting DynDNS bridge for FritzBox -> Hetzner DNS
Starting DynDNS server on port 8080
Update URL: http://localhost:8080/update?hostname=yourdomain.com&myip=1.2.3.4
Configure your FritzBox with:
  Update URL: http://your-server:8080/update
  Domain: <your-domain>
  Username: admin
  Password: your_secure_password
```

## FritzBox Configuration

Configure your FritzBox for dynamic DNS:

1. Go to **Internet** → **Permit Access** → **DynDNS**
2. Select **DynDNS Provider**: User-defined
3. **Update URL**: `http://your-server-ip:8080/update?hostname=<domain>&myip=<ipaddr>`
4. **Domain name**: Your domain or subdomain (e.g., `home.yourdomain.com`)
5. **Username**: `admin` (or your custom username)
6. **Password**: Your `DYNDNS_PASSWORD`

## API Usage Examples

### Update IPv4 Record
```bash
curl -u admin:password "http://localhost:8080/update?hostname=home.example.com&myip=203.0.113.1"
```

### Update IPv6 Record
```bash
curl -u admin:password "http://localhost:8080/update?hostname=home.example.com&myipv6=2001:db8::1"
```

### Update Both IPv4 and IPv6 (Dual-Stack)
```bash
curl -u admin:password "http://localhost:8080/update?hostname=home.example.com&myip=203.0.113.1&myipv6=2001:db8::1"
```

### Auto-detect Client IP
```bash
curl -u admin:password "http://localhost:8080/update?hostname=home.example.com"
```

## Response Format

The server returns FritzBox-compatible responses:

- **Success**: `good 203.0.113.1` or `good IPv4: 203.0.113.1, IPv6: 2001:db8::1`
- **Error**: `911` (general error)
- **Offline**: `good` (for offline requests)

## Supported DNS Record Types

The Hetzner DNS API client supports all standard DNS record types:

- **A** - IPv4 address records
- **AAAA** - IPv6 address records  
- **CNAME** - Canonical name records
- **MX** - Mail exchange records
- **TXT** - Text records
- **NS** - Name server records
- **PTR** - Pointer records
- **SRV** - Service records

## Direct API Usage

You can also use the Hetzner DNS client directly in your Go code:

```go
package main

import (
    "fmt"
    "log"
)

func main() {
    // Create client
    client := NewClient("your-api-key")

    // Get all zones
    zones, err := client.GetZones()
    if err != nil {
        log.Fatal(err)
    }

    // Create a new A record
    ttl := 3600
    record, err := client.CreateRecord(CreateRecordRequest{
        Type:   "A",
        Name:   "test",
        Value:  "192.168.1.1",
        TTL:    &ttl,
        ZoneID: zones[0].ID,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Created record: %+v\n", record)

    // Update the record
    updatedRecord, err := client.UpdateRecord(record.ID, UpdateRecordRequest{
        Type:  "A",
        Name:  "test",
        Value: "192.168.1.2",
        TTL:   &ttl,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Updated record: %+v\n", updatedRecord)

    // Delete the record
    err = client.DeleteRecord(record.ID)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Record deleted")
}
```

## Architecture

```
┌─────────────┐    HTTP POST     ┌──────────────────┐    REST API    ┌─────────────────┐
│   FritzBox  │ ──────────────→  │   DynDNS Bridge  │ ─────────────→ │  Hetzner DNS    │
│             │  Basic Auth      │                  │   API Token    │     API         │
└─────────────┘                  └──────────────────┘                └─────────────────┘
                                          │
                                          ▼
                                  ┌──────────────────┐
                                  │  DNS Records     │
                                  │  - A Records     │
                                  │  - AAAA Records  │
                                  │  - Auto-create   │
                                  │  - Auto-update   │
                                  └──────────────────┘
```

## Testing

The project includes comprehensive tests covering all functionality:

```bash
# Run all tests
go test -v

# Run specific test files
go test -v -run TestDynDNS
go test -v -run TestClient
go test -v -run TestTypes
```

Test coverage includes:
- ✅ HTTP client functionality
- ✅ DNS record CRUD operations
- ✅ DynDNS server endpoints
- ✅ Authentication mechanisms
- ✅ IPv4/IPv6 validation
- ✅ JSON marshaling/unmarshaling
- ✅ Error handling scenarios
‚
## Deployment

### Docker Multi-Architecture Build

The project supports building Docker images for multiple architectures (AMD64 and ARM64).

#### Single Architecture Build

Build for your current platform:
```bash
docker build -t hetzner-dns-bridge .
```

#### Multi-Architecture Build

Build for both AMD64 and ARM64:
```bash
# Create and use a new builder that supports multi-platform builds
docker buildx create --name multiarch --driver docker-container --use
docker buildx inspect --bootstrap

# Build and push multi-arch images
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t yourusername/hetzner-dns-bridge:latest \
  --push .
```

#### Architecture-Specific Builds

Build for AMD64 only:
```bash
docker buildx build \
  --platform linux/amd64 \
  -t hetzner-dns-bridge:amd64 \
  --load .
```

Build for ARM64 only:
```bash
docker buildx build \
  --platform linux/arm64 \
  -t hetzner-dns-bridge:arm64 \
  --load .
```

#### Running the Container

```bash
# Basic run
docker run -p 8080:8080 \
  -e HETZNER_DNS_API_KEY="your-token" \
  -e DYNDNS_PASSWORD="your-password" \
  hetzner-dns-bridge

# With all environment variables
docker run -d \
  --name hetzner-dns-bridge \
  --restart unless-stopped \
  -p 8080:8080 \
  -e HETZNER_DNS_API_KEY="your-token" \
  -e DYNDNS_PASSWORD="your-password" \
  -e DYNDNS_USERNAME="admin" \
  -e DYNDNS_PORT="8080" \
  hetzner-dns-bridge
```

#### Docker Compose

Create a `docker-compose.yml` file:

```yaml
version: '3.8'

services:
  hetzner-dns-bridge:
    build: .
    # Or use pre-built image:
    # image: yourusername/hetzner-dns-bridge:latest
    container_name: hetzner-dns-bridge
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - HETZNER_DNS_API_KEY=${HETZNER_DNS_API_KEY}
      - DYNDNS_PASSWORD=${DYNDNS_PASSWORD}
      - DYNDNS_USERNAME=${DYNDNS_USERNAME:-admin}
      - DYNDNS_PORT=${DYNDNS_PORT:-8080}
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

Run with Docker Compose:
```bash
# Create .env file with your credentials
echo "HETZNER_DNS_API_KEY=your-token" > .env
echo "DYNDNS_PASSWORD=your-password" >> .env

# Start the service
docker-compose up -d

# View logs
docker-compose logs -f
```

### Kubernetes Deployment

Create `k8s-deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hetzner-dns-bridge
  labels:
    app: hetzner-dns-bridge
spec:
  replicas: 1
  selector:
    matchLabels:
      app: hetzner-dns-bridge
  template:
    metadata:
      labels:
        app: hetzner-dns-bridge
    spec:
      containers:
      - name: hetzner-dns-bridge
        image: yourusername/hetzner-dns-bridge:latest
        ports:
        - containerPort: 8080
        env:
        - name: HETZNER_DNS_API_KEY
          valueFrom:
            secretKeyRef:
              name: hetzner-dns-secrets
              key: api-key
        - name: DYNDNS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: hetzner-dns-secrets
              key: dyndns-password
        - name: DYNDNS_USERNAME
          value: "admin"
        - name: DYNDNS_PORT
          value: "8080"
        livenessProbe:
          httpGet:
            path: /
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
        resources:
          requests:
            memory: "32Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "200m"
        securityContext:
          allowPrivilegeEscalation: false
          runAsNonRoot: true
          runAsUser: 1001
          readOnlyRootFilesystem: true
          capabilities:
            drop:
            - ALL

---
apiVersion: v1
kind: Service
metadata:
  name: hetzner-dns-bridge-service
spec:
  selector:
    app: hetzner-dns-bridge
  ports:
    - protocol: TCP
      port: 8080
      targetPort: 8080
  type: ClusterIP

---
apiVersion: v1
kind: Secret
metadata:
  name: hetzner-dns-secrets
type: Opaque
stringData:
  api-key: "your-hetzner-api-token"
  dyndns-password: "your-dyndns-password"
```

Deploy to Kubernetes:
```bash
kubectl apply -f k8s-deployment.yaml
```
