# Headless Print Service

A high-performance HTTP service that generates PDFs from HTML content using headless Chrome with optimized resource usage.

## Features

- HTTP endpoint for PDF generation
- Browser tab pooling for efficient resource usage
- Dockerized application with separate Chrome and application containers
- Optimized Chrome flags for headless operation in containerized environments

## Prerequisites

- Docker and Docker Compose (for containerized deployment)
- Go 1.24+ (for local development)

## Running with Docker

The easiest way to run the application is using Docker Compose:

```bash
# Build and start the service
docker compose up -d

# To stop the service
docker compose down
```

The service will be available at http://localhost:8080

The setup includes two containers:
- `chrome`: A headless Chrome instance using the `chromedp/headless-shell` image
- `pdf-generator`: The Go application that connects to Chrome for PDF generation

## Running Locally

To run the application locally:

1. Make sure you have Chrome or Chromium installed
2. Run the application with default settings (using local Chrome):

```bash
go run main.go
```

Or connect to a remote Chrome DevTools Protocol (CDP) server:

```bash
go run main.go --use-cdp-server --url ws://localhost:9222
```

The service will be available at http://localhost:8080

## API Endpoints

- `GET /generate-pdf` - Generates a PDF from the HTML template and returns it as a download

## Customizing the PDF

To customize the PDF content, modify the `index.html` file. The application reads this file and converts it to a PDF.

## Building the Docker Image Manually

If you want to build and run the Docker image manually:

```bash
# Build the image
docker build -t headless-print .

# Run the container (make sure Chrome container is running first)
docker run -p 8080:8080 -v $(pwd)/index.html:/app/index.html headless-print
```

## Chrome Configuration

The application uses the following Chrome flags for optimal performance in containerized environments:

```
--no-sandbox
--disable-setuid-sandbox
--no-zygote
--disable-dev-shm-usage
--disable-gpu
--disable-extensions
--disable-plugins
--disable-background-timer-throttling
--disable-backgrounding-occluded-windows
--disable-renderer-backgrounding
--disable-web-security
--disable-software-rasterizer
--disable-default-apps
--disable-component-extensions-with-background-pages
--disable-sync
--headless
```

## Resource Management

The application implements a tab pool to efficiently manage Chrome tabs and optimize resource usage. The pool size is automatically calculated based on available CPU resources.
