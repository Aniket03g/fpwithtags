# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd/mcp-bridge/ ./cmd/mcp-bridge/

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o mcp-bridge ./cmd/mcp-bridge

# Runtime stage
FROM alpine:latest

# Install Docker CLI (needed to spawn MCP containers)
RUN apk --no-cache add docker-cli

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/mcp-bridge .

# Expose port
EXPOSE 8089

# Run the bridge
CMD ["./mcp-bridge"]
