# syntax=docker.io/docker/dockerfile:1@sha256:443aab4ca21183e069e7d8b2dc68006594f40bddf1b15bbd83f5137bd93e80e2

FROM --platform=$BUILDPLATFORM gcr.io/distroless/static:nonroot@sha256:2556293984c5738fc75208cce52cf0a4762c709cf38e4bf8def65a61992da0ad AS distroless
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1@sha256:800d9b4fb6231053473df14d5a7116bfd33500bca5ca4c6d544de739d9a7d302 AS golang

FROM golang AS builder
WORKDIR /app
ENV CGO_ENABLED=0
COPY go.??? .
RUN \
  --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
    set -ux \
 && go mod download
COPY . .
RUN \
  --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
    set -ux \
 && GOOS=linux GOARCH=amd64 go build -mod=readonly -o srv -a -installsuffix cgo -ldflags '-w -s -extldflags "-static"' ./cmd/


FROM distroless AS srv
COPY --from=builder --chown=nonroot /app/srv /srv
ENTRYPOINT ["/srv"]
