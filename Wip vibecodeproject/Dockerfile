# Multi-stage build for multiple architectures
FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build arguments for cross-compilation
ARG TARGETOS
ARG TARGETARCH

# Build the application
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -a -installsuffix cgo -o hetzner-dns-bridge .

# Final stage - minimal runtime image
FROM --platform=$TARGETPLATFORM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user for security
RUN addgroup -g 1001 -S hetzner && \
    adduser -u 1001 -S hetzner -G hetzner

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/hetzner-dns-bridge .

# Change ownership to non-root user
RUN chown hetzner:hetzner /app/hetzner-dns-bridge

# Switch to non-root user
USER hetzner

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Set default environment variables
ENV DYNDNS_PORT=8080
ENV DYNDNS_USERNAME=admin

# Run the application
CMD ["./hetzner-dns-bridge"]
