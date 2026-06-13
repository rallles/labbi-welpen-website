# ---- Build Stage ----
FROM golang:1.24.1-alpine AS builder

WORKDIR /app

# Cache go.mod & go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o labbi-app ./cmd

# ---- Final Stage ----
FROM alpine:3.21.3

WORKDIR /app

# Install CA certificates for TLS
RUN apk add --no-cache ca-certificates && mkdir -p /app/data/uploads

# Copy the binary
COPY --from=builder /app/labbi-app ./labbi-app

# Copy public assets, templates, and crawler metadata
COPY --from=builder /app/static ./static
COPY --from=builder /app/internal/templates ./templates
COPY --from=builder /app/robots.txt ./robots.txt
COPY --from=builder /app/sitemap.xml ./sitemap.xml

# Expose port
EXPOSE 8080

# Default environment (can be overridden)
ENV SERVER_ADDRESS=":8080"
ENV UPLOAD_DIR="/app/data/uploads"
ENV STATIC_DIR="/app/static"
ENV TEMPLATE_DIR="/app/templates"

# Entry point
ENTRYPOINT ["/app/labbi-app"]
