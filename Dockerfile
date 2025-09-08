FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Copy source code
COPY . .

# Build the application
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o app .

# Use Alpine Chrome as the base image for the final stage
FROM zenika/alpine-chrome:with-node

# Note: The zenika/alpine-chrome image already has Chrome installed and runs as a non-root user

# Create app directory and set permissions
USER root
RUN mkdir -p /app && chown -R chrome:chrome /app

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/app .

# Copy the HTML template
COPY index.html .

# Set correct permissions
RUN chmod +x /app/app && chown -R chrome:chrome /app

# Switch back to non-root user
USER chrome

# Expose the port the app runs on
EXPOSE 8080

# Set Chrome environment variables
ENV CHROME_BIN=/usr/bin/chromium-browser \
    CHROME_PATH=/usr/lib/chromium/ \
    PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=true \
    PUPPETEER_EXECUTABLE_PATH=/usr/bin/chromium-browser

# Command to run the application with Chrome flags
CMD ["./app", "--use-cdp-server", "--url", "ws://chrome:9222"]
