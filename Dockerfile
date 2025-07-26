# * for amd64: docker build -t ppcelery/one-api:amd64-latest .
# * for arm64: DOCKER_BUILDKIT=1 docker build --platform linux/arm64 --build-arg TARGETARCH=arm64 -t ppcelery/one-api:arm64-latest .
FROM node:24-bookworm AS builder

RUN npm install -g npm react-scripts

WORKDIR /web
COPY ./VERSION .
COPY ./web .

# Install dependencies for each project
# do not build parallel to avoid OOM on github actions
RUN cd /web/default && yarn install
RUN cd /web/berry && yarn install
RUN cd /web/air && yarn install

RUN mkdir -p /web/build

# Build the web projects
# do not build parallel to avoid OOM on github actions
RUN DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$(cat ./VERSION) npm run build --prefix /web/default
RUN DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$(cat ./VERSION) npm run build --prefix /web/berry
RUN DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$(cat ./VERSION) npm run build --prefix /web/air

FROM golang:1.24.4-bookworm AS builder2

# Make sure to use ARG with a default value
ARG TARGETARCH=amd64

# Set proper environment variables based on TARGETARCH
ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=${TARGETARCH}

# Print architecture information for debugging
RUN echo "Building for TARGETARCH=${TARGETARCH}" && \
    echo "Current architecture: $(uname -m)"

# For ARM64 builds
RUN apt-get update && \
    if [ "${TARGETARCH}" = "arm64" ]; then \
        apt-get install -y gcc-aarch64-linux-gnu && \
        export CC=aarch64-linux-gnu-gcc && \
        export GOARCH=arm64 && \
        export CGO_ENABLED=1 && \
        # This is critical for ARM64 cross-compilation
        export CGO_CFLAGS="-g -O2 -fPIC"; \
    else \
        apt-get install -y build-essential; \
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

# Simplified build command that handles both architectures
RUN if [ "${TARGETARCH}" = "arm64" ]; then \
        CC=aarch64-linux-gnu-gcc \
        CGO_ENABLED=1 \
        GOARCH=arm64 \
        CGO_CFLAGS="-g -O2 -fPIC" \
        go build -trimpath -ldflags "-s -w -X github.com/songquanpeng/one-api/common.Version=$(cat VERSION)" -o one-api; \
    else \
        go build -trimpath -ldflags "-s -w -X github.com/songquanpeng/one-api/common.Version=$(cat VERSION)" -o one-api; \
    fi

# Use Fedora as the base image which has better ARM64 support
FROM fedora:latest

ARG TARGETARCH=amd64
ENV NO_AT_BRIDGE=1

# Install basic requirements
RUN dnf install -y --setopt=install_weak_deps=False \
    ca-certificates tzdata bash haveged && \
    dnf clean all

# Create a non-root user with fixed name and password
RUN useradd -m -s /bin/bash appuser && \
    echo "appuser:hwe@ecn.ynr0VZK_kwk" | chpasswd && \
    mkdir -p /data && chown appuser:appuser /data

# Copy the one-api binary
COPY --from=builder2 /build/one-api /usr/local/bin/one-api

# Switch to non-root user
USER appuser

EXPOSE 3000
WORKDIR /data
ENTRYPOINT ["/usr/local/bin/one-api"]