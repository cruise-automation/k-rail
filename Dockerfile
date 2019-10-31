# Build stage
ARG GO_VERSION=1.13
FROM golang:${GO_VERSION}-alpine AS builder
RUN mkdir build; apk --no-cache add ca-certificates git make
WORKDIR /go/src/github.com/cruise-automation/k-rail
COPY ./ /build/
RUN cd /build; make test; make build

# Production image build stage
FROM scratch
EXPOSE 8443/tcp
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /build/k-rail /k-rail
USER 65534
ENTRYPOINT ["/k-rail", "-config", "/config/config.yml"]
