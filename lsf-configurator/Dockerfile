# Build stage: use Debian-based Go image for full CGO support
FROM golang:1.23.2-bullseye AS build

RUN apt-get update && apt-get install -y build-essential git && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY . /app
RUN CGO_ENABLED=1 GOOS=linux go build -a -o lsf-configurator

# Build stage for React frontend
FROM node:22-alpine AS frontend-build
WORKDIR /frontend
COPY ./frontend /frontend
RUN npm install && npm run build

# Final image: minimal Debian-based Go image with CGO support
FROM golang:1.23.2-bullseye

RUN apt-get update && apt-get install -y build-essential && rm -rf /var/lib/apt/lists/*
WORKDIR /
COPY --from=build /app/lsf-configurator /lsf-configurator
COPY --from=frontend-build /frontend/dist /public
COPY --from=build /app/templates /templates

EXPOSE 8080
CMD ["/lsf-configurator"]
