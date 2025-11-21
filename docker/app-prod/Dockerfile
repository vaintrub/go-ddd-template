# Build stage
FROM golang:1.25-alpine AS builder

ARG SERVICE

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire internal directory
COPY internal/ ./internal/

# Build the specific service
WORKDIR /build/internal/${SERVICE}
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o /app .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS connections
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /home/appuser

# Copy the binary from builder
COPY --from=builder /app ./app

# Use non-root user
USER appuser

# Run the application
CMD ["./app"]
