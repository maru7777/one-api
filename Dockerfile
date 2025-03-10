FROM node:18 AS builder

WORKDIR /web
COPY ./VERSION .
COPY ./web .

# Fix the React build issues by installing dependencies globally first
RUN npm install -g react-scripts

# Install dependencies for each project
RUN npm install --prefix /web/default & \
    npm install --prefix /web/berry & \
    npm install --prefix /web/air & \
    wait

# Build the web projects
RUN DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$(cat ./VERSION) npm run build --prefix /web/default & \
    DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$(cat ./VERSION) npm run build --prefix /web/berry & \
    DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$(cat ./VERSION) npm run build --prefix /web/air & \
    wait

FROM golang:1.24.1-bullseye AS builder2

# Make sure to use ARG with a default value
ARG TARGETARCH=amd64

# Set proper environment variables based on TARGETARCH
ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=${TARGETARCH}

# For ARM64 builds
RUN if [ "${TARGETARCH}" = "arm64" ]; then \
        apt-get update && apt-get install -y gcc-aarch64-linux-gnu && \
        export CC=aarch64-linux-gnu-gcc; \
    else \
        apt-get update && apt-get install -y build-essential; \
    fi

# Common dependencies
RUN apt-get install -y --no-install-recommends \
    sqlite3 libsqlite3-dev && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=builder /web/build ./web/build

# Use the appropriate compiler based on architecture
RUN if [ "${TARGETARCH}" = "arm64" ]; then \
        CC=aarch64-linux-gnu-gcc go build -trimpath -ldflags "-s -w -X github.com/songquanpeng/one-api/common.Version=$(cat VERSION)" -o one-api; \
    else \
        go build -trimpath -ldflags "-s -w -X github.com/songquanpeng/one-api/common.Version=$(cat VERSION)" -o one-api; \
    fi

# Use a pre-built image that already has ffmpeg for ARM64
FROM --platform=$TARGETPLATFORM jrottenberg/ffmpeg:4.3-ubuntu2004 AS ffmpeg

# Use Ubuntu as the base image which has better ARM64 support
FROM --platform=$TARGETPLATFORM ubuntu:20.04

ARG TARGETARCH=amd64
ENV DEBIAN_FRONTEND=noninteractive

# Install basic requirements without triggering libc-bin reconfiguration
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates tzdata bash haveged && \
    rm -rf /var/lib/apt/lists/*

# Copy ffmpeg binaries from the ffmpeg image
COPY --from=ffmpeg /usr/local/bin/ffmpeg /usr/local/bin/
COPY --from=ffmpeg /usr/local/bin/ffprobe /usr/local/bin/

# Copy our application binary
COPY --from=builder2 /build/one-api /

EXPOSE 3000
WORKDIR /data
ENTRYPOINT ["/one-api"]
