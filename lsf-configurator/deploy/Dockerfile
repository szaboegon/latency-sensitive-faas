# Build stage: use Debian-based Go image for full CGO support
FROM golang:1.23.2-bullseye AS build

RUN apt-get update && apt-get install -y build-essential git && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY . /app
RUN CGO_ENABLED=1 GOOS=linux go build -a -o lsf-configurator

# Final image: minimal Debian-based Go image with CGO support
FROM golang:1.23.2-bullseye

RUN apt-get update && apt-get install -y build-essential && rm -rf /var/lib/apt/lists/*
WORKDIR /
COPY --from=build /app/lsf-configurator /lsf-configurator
COPY --from=build /app/public /public
COPY --from=build /app/templates /templates

EXPOSE 8080
CMD ["/lsf-configurator"]
