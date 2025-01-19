FROM node:18 as builder

WORKDIR /web
COPY ./VERSION .
COPY ./web .

WORKDIR /web/default
RUN npm install
RUN DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$(cat ../VERSION) npm run build

WORKDIR /web/berry
RUN npm install
RUN DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$(cat ../VERSION) npm run build

WORKDIR /web/air
RUN npm install
RUN DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$(cat ../VERSION) npm run build

FROM golang:1.23.5-bullseye AS builder2

RUN apt-get update
RUN apt-get install -y --no-install-recommends g++ make gcc git build-essential ca-certificates \
    && update-ca-certificates 2>/dev/null || true \
    && rm -rf /var/lib/apt/lists/*

ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux

WORKDIR /build
ADD go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=builder /web/build ./web/build
RUN go build -trimpath -ldflags "-s -w -X 'github.com/songquanpeng/one-api/common.Version=$(cat VERSION)' -extldflags '-static'" -o one-api

FROM debian:bullseye

RUN apt-get update
RUN apt-get install -y --no-install-recommends ca-certificates haveged tzdata ffmpeg \
    && update-ca-certificates 2>/dev/null || true \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder2 /build/one-api /
EXPOSE 3000
WORKDIR /data
ENTRYPOINT ["/one-api"]
