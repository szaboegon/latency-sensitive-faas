FROM docker.io/golang:latest as build
# Copy the source files from the host
COPY . /
# Set the working directory to the same place we copied the code
WORKDIR /
# Build a binary named 'lsf-configurator'
RUN CGO_ENABLED=0 GOOS=linux go build -a -o lsf-configurator

# Stage 2
FROM alpine:latest
# Install Git for cloning template repos
RUN apk add --no-cache git
# Copy the binary from the host and other necessary files
COPY --from=build /lsf-configurator .
COPY --from=build ./public /public
COPY --from=build ./templates /templates
# Expose port on which the server listens
EXPOSE 8080
# Run the built binary
CMD ["/lsf-configurator"]