FROM node:18 as builder

WORKDIR /web
COPY ./VERSION .
COPY ./web .

WORKDIR /web/default
RUN npm install
RUN DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$(cat VERSION) npm run build

WORKDIR /web/berry
RUN npm install
RUN DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$(cat VERSION) npm run build

FROM golang:1.22.0-bullseye AS builder2

ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux

WORKDIR /build
ADD go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=builder /web/build ./web/build
RUN go build -ldflags "-s -w -X 'one-api/common.Version=$(cat VERSION)' -extldflags '-static'" -o one-api

FROM debian:bullseye

RUN apt-get update
RUN apt-get install -y --no-install-recommends ca-certificates haveged tzdata \
    # for google-chrome
    # libappindicator1 fonts-liberation xdg-utils wget \
    # libasound2 libatk-bridge2.0-0 libatspi2.0-0 libcurl3-gnutls libcurl3-nss \
    # libcurl4 libcurl3 libdrm2 libgbm1 libgtk-3-0 libgtk-4-1 libnspr4 libnss3 \
    # libu2f-udev libvulkan1 libxkbcommon0 \
    && update-ca-certificates 2>/dev/null || true \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder2 /build/one-api /
EXPOSE 3000
WORKDIR /data
ENTRYPOINT ["/one-api"]
