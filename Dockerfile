# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git build-base

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum from backend (only one module)
COPY backend/go.mod backend/go.sum ./backend/
# Copy pkg/featureplus go.mod so "replace" works
COPY pkg/featureplus/go.mod ./pkg/featureplus/

# Download dependencies
RUN cd backend && go mod download

# Copy source code (backend + pkg)
COPY backend ./backend
COPY pkg ./pkg

# Build the application
WORKDIR /app/backend
RUN go mod tidy
RUN CGO_ENABLED=1 GOOS=linux go build -o featureplus-server main.go


# Runtime stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/backend/featureplus-server .

# Copy templates and static files into runtime container
COPY backend/templates ./templates
COPY backend/static ./static

# Expose the port your app runs on
EXPOSE 8080

CMD ["./featureplus-server"]