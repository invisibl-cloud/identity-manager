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

FROM gcr.io/distroless/static-debian11:nonroot
WORKDIR /app
ENTRYPOINT ["/app/identity-manager"]
