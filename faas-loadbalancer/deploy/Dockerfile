FROM docker.io/golang:latest as build
# Copy the source files from the host
COPY . /
# Set the working directory to the same place we copied the code
WORKDIR /
# Build a binary named 'splitdim'
RUN CGO_ENABLED=0 GOOS=linux go build -a -o faas-loadbalancer

# Stage 2
FROM scratch
# Copy the binary from the host
COPY --from=build /faas-loadbalancer .

# Expose port on which the server listens
EXPOSE 8080
# Run the built binary
CMD ["/faas-loadbalancer"]