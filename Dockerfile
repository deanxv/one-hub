FROM node:18 as builder

WORKDIR /build

COPY web/package.json .
COPY web/yarn.lock .

RUN yarn config set registry https://registry.npmmirror.com && \
    yarn config set network-timeout 600000 && \
    yarn config set network-concurrency 1 && \
    yarn --frozen-lockfile --network-timeout 600000

COPY ./web .
COPY ./VERSION .
RUN DISABLE_ESLINT_PLUGIN='true' VITE_APP_VERSION=$(cat VERSION) npm run build

FROM golang:1.24.2 AS builder2

ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOPROXY=https://goproxy.cn,direct

WORKDIR /build
ADD go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=builder /build/build ./web/build
RUN go build -ldflags "-s -w -X 'one-api/common.Version=$(cat VERSION)' -extldflags '-static'" -o done-hub

FROM alpine

RUN apk update \
    && apk upgrade \
    && apk add --no-cache ca-certificates tzdata \
    && update-ca-certificates 2>/dev/null || true

COPY --from=builder2 /build/done-hub /
EXPOSE 3000
WORKDIR /data
ENTRYPOINT ["/done-hub"]
