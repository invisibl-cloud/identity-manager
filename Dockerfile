FROM golang as base

FROM scratch
ARG TARGETARCH
ARG TARGETOS
COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=base /usr/share/zoneinfo /usr/share/zoneinfo
WORKDIR /app
COPY dist/identity-manager_${TARGETOS}_$TARGETARCH /app/identity-manager
ENTRYPOINT ["/app/identity-manager"]
