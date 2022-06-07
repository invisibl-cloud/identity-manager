FROM golang:1.17.11 as builder
WORKDIR /app
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT=""
ARG GOPROXY
ARG BUILD_DATE
ARG COMMIT_HASH
ARG VERSION
ARG CGO_ENABLED="0"
ENV GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOARM=${TARGETVARIANT}
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make release-binary

FROM alpine:3.16.0
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/identity-manager /app/identity-manager
WORKDIR /app
ENTRYPOINT ["/app/identity-manager"]
