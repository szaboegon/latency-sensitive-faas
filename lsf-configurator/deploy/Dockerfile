# Use a Debian-based Golang image for building the binary with full CGO support
FROM golang:1.23.2-bullseye AS build

# Install build tools
RUN apt-get update && apt-get install -y build-essential git && rm -rf /var/lib/apt/lists/*

# Copy the source files from the host
COPY . /app

# Set the working directory
WORKDIR /app

# Build the Go binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -o lsf-configurator

# Minimal runtime image with CGO support (Debian-based)
FROM golang:1.23.2-bullseye

# Install necessary packages
RUN apt-get update && apt-get install -y build-essential && rm -rf /var/lib/apt/lists/*

# Copy the built binary and necessary directories from the build stage
COPY --from=build /app/lsf-configurator /lsf-configurator
COPY --from=build /app/public /public
COPY --from=build /app/templates /templates

# Expose port 8080 for the server
EXPOSE 8080

# Run the built binary
CMD ["/lsf-configurator"]
