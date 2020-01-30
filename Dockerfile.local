FROM alpine as builder
RUN apk add --no-cache ca-certificates

FROM alpine
EXPOSE 8443/tcp
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY k-rail /k-rail
USER 65534
ENTRYPOINT ["/k-rail", "-config", "/config/config.yml"]