# PDF Generator Service

A simple HTTP service that generates PDFs from HTML content using headless Chrome.

## Features

- HTTP endpoint for PDF generation
- Simple web interface
- Dockerized application for easy deployment

## Prerequisites

- Docker and Docker Compose (for containerized deployment)
- Go 1.16+ (for local development)

## Running with Docker

The easiest way to run the application is using Docker Compose:

```bash
# Build and start the service
docker-compose up -d

# To stop the service
docker-compose down
```

The service will be available at http://localhost:8080

## Running Locally

To run the application locally:

1. Make sure you have Chrome or Chromium installed
2. Run the application:

```bash
go run main.go
```

The service will be available at http://localhost:8080

## API Endpoints

- `GET /` - Home page with a button to generate PDF
- `GET /generate-pdf` - Generates a PDF from the HTML template and returns it as a download

## Customizing the PDF

To customize the PDF content, modify the `index.html` file. The application reads this file and converts it to a PDF.

## Building the Docker Image Manually

If you want to build and run the Docker image manually:

```bash
# Build the image
docker build -t lightpanda-pdf .

# Run the container
docker run -p 8080:8080 -v $(pwd)/index.html:/app/index.html lightpanda-pdf
```
