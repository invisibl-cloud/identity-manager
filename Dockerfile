# Build
FROM golang:1.21.0-alpine3.18 AS builder
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT=""
ARG APP_VERSION
ARG COMMIT_HASH
ARG GIT_REF
ARG BUILD_DATE
ARG BUILD_BY=docker
ARG GOPROXY
RUN apk add --no-cache --update ca-certificates git
ENV CGO_ENABLED=0 GO111MODULE=on GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOARM=${TARGETVARIANT}
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags "-s -w -X main.version=${APP_VERSION} -X main.commit=${COMMIT_HASH} -X main.date=${BUILD_DATE} -linkmode internal -extldflags -static" -o identity-manager main.go

# Image
FROM gcr.io/distroless/static-debian11:nonroot
#COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
# Copy module files for CVE scanning / dependency analysis.
COPY --from=builder /app/go.mod /app/go.sum /app/
COPY --from=builder /app/identity-manager /app/
ENTRYPOINT ["/app/identity-manager"]
