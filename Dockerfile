FROM --platform=$BUILDPLATFORM golang:alpine AS build
ARG TARGETOS
ARG TARGETARCH
WORKDIR /app
COPY . ./
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o fritzbox-hetzner-dyndns .

FROM alpine
COPY --from=build /app/fritzbox-hetzner-dyndns /fritzbox-hetzner-dyndns

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1
    
ENTRYPOINT ["/fritzbox-hetzner-dyndns"]