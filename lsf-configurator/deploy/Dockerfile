# Use a Golang image for building the binary
FROM docker.io/golang:latest AS build
# Copy the source files from the host
COPY . /app
# Set the working directory
WORKDIR /app
# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -o lsf-configurator

# Stage 2: Use an image with Podman preinstalled
FROM quay.io/containers/podman:latest
# Install Git
RUN dnf install -y git && dnf clean all
# Copy the built binary and necessary directories
COPY --from=build /app/lsf-configurator /lsf-configurator
COPY --from=build /app/public /public
COPY --from=build /app/templates /templates
# Expose port 8080 for the server
EXPOSE 8080
# Run the built binary
CMD ["/lsf-configurator"]
