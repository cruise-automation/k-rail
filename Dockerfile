# Build stage
ARG GO_VERSION=1.16
FROM golang:${GO_VERSION}-buster AS builder
RUN apt-get update && \
  apt-get -y install protobuf-compiler && \
  apt-get clean && \
  rm -rf /var/lib/apt/lists/* && \
  rm -rf /var/cache/apt/* && \
  protoc --version
RUN go get -u github.com/golang/protobuf/protoc-gen-go google.golang.org/grpc/cmd/protoc-gen-go-grpc

WORKDIR /build
COPY ./ /build/
RUN make build
RUN make test

# Production image build stage
FROM scratch
EXPOSE 8443/tcp
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /build/k-rail /k-rail
COPY --from=builder /build/k-rail /k-rail-check
USER 65534
ENTRYPOINT ["/k-rail", "-config", "/config/config.yml"]
