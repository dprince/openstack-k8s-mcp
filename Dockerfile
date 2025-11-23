# Build stage
FROM golang:1.24.4 AS builder

WORKDIR /workspace

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY cmd/ cmd/
COPY internal/ internal/

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o openstack-k8s-mcp ./cmd/openstack-k8s-mcp

# Final stage
FROM gcr.io/distroless/static:nonroot

WORKDIR /

# Copy the binary from builder
COPY --from=builder /workspace/openstack-k8s-mcp /openstack-k8s-mcp

USER 65532:65532

ENTRYPOINT ["/openstack-k8s-mcp"]
