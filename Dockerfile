FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o wireguard-manager

# Create final image
FROM alpine:latest

# Install required packages
RUN apk add --no-cache wireguard-tools

# Copy the binary from builder
COPY --from=builder /app/wireguard-manager /usr/local/bin/
COPY wireguard-install.sh /usr/local/bin/

# Make the script executable
RUN chmod +x /usr/local/bin/wireguard-install.sh

# Set environment variables
ENV WG_AUTH_TOKEN=your-secure-token

# Expose port
EXPOSE 8080

# Run the application
CMD ["wireguard-manager"] 