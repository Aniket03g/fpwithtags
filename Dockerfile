# Build stage
FROM golang:1.20-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY backend/go.mod backend/go.sum ./backend/
COPY pkg/featureplus/go.mod pkg/featureplus/go.sum ./pkg/featureplus/

# Download dependencies for both modules
RUN cd backend && go mod download
RUN cd pkg/featureplus && go mod download

# Copy the source code
COPY backend ./backend
COPY pkg ./pkg

# Build the application
WORKDIR /app/backend
RUN CGO_ENABLED=1 GOOS=linux go build -o featureplus-server main.go

# Runtime stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata sqlite

# Create a non-root user
RUN adduser -D -u 1000 appuser

# Create data directory and set permissions
RUN mkdir -p /data && chown -R appuser:appuser /data

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/backend/featureplus-server .

# Copy static assets and templates
COPY backend/static ./static
COPY backend/templates ./templates

# Set environment variables
ENV FP_DB_PATH=/data/featureplus.sqlite
ENV GIN_MODE=release

# Expose the port
EXPOSE 8080

# Use non-root user
USER appuser

# Declare volume for data persistence
VOLUME ["/data"]

# Run the application
CMD ["./featureplus-server"]
