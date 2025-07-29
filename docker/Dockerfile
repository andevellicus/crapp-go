# docker/Dockerfile

# --- Build Stage ---
# Use the official Go image as a builder, matching your go.mod version.
FROM golang:1.24.2-alpine AS builder

# Install air for live reloading
RUN go install github.com/air-verse/air@latest

# Set the working directory inside the container.
WORKDIR /app

# Copy the Go module files.
COPY server/go.mod server/go.sum ./
RUN go mod download

# Copy the rest of the application source code.
COPY server/ ./
COPY config/ ./config
COPY tailwind.config.js ./
COPY postcss.config.js ./

# Generate the templ files.
RUN go install github.com/a-h/templ/cmd/templ@latest
RUN templ generate

# Build the Go application.
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main .

# --- Final Stage ---
# Use a minimal base image for the final container.
FROM alpine:latest

WORKDIR /app

# Copy the compiled application binary and air from the builder stage.
COPY --from=builder /go/bin/air /usr/local/bin/
COPY --from=builder /app/main .

# Copy the static assets and configuration files needed at runtime.
COPY --from=builder /app/assets ./assets
COPY --from=builder /app/config ./config

# Expose the port the application will run on.
EXPOSE 8080

# The command to run when the container starts.
CMD ["/app/main"]