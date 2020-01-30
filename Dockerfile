# Build stage
ARG GO_VERSION=1.13
FROM golang:${GO_VERSION}-buster AS builder
WORKDIR /build
COPY ./ /build/
RUN make test
RUN make build

# Production image build stage
FROM scratch
EXPOSE 8443/tcp
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /build/k-rail /k-rail
USER 65534
ENTRYPOINT ["/k-rail", "-config", "/config/config.yml"]
