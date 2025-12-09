# Build stage
FROM golang:1.25 AS builder

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    ca-certificates \
    tzdata \
  && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /build

# Copy go mod files for better layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary with optimizations
# CGO_DISABLED to keep the binary portable and lean
# -ldflags="-w -s" to strip debug info and reduce binary size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o tchat

# Final stage
FROM ubuntu:24.04

# Install runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
  && rm -rf /var/lib/apt/lists/*

# Create non-root user for security (fallback if UID 1000 already exists)
RUN if id -u 1000 >/dev/null 2>&1; then \
      useradd -m -s /bin/bash appuser; \
    else \
      useradd -m -u 1000 -s /bin/bash appuser; \
    fi

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder --chown=appuser:appuser /build/tchat .

# Switch to non-root user
USER appuser

# Expose port (configurable via PORT env var, default 8080)
EXPOSE 8080

# Run the application
CMD ["./tchat"]
