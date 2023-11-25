FROM golang:1.21-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /external-dns-nextdns-webhook

FROM alpine:3

WORKDIR /app

COPY --from=builder /external-dns-nextdns-webhook /external-dns-nextdns-webhook

EXPOSE 8888

ENTRYPOINT ["/external-dns-nextdns-webhook"]

LABEL org.opencontainers.image.title="ExternalDNS NextDNS webhook Docker Image" \
      org.opencontainers.image.description="external-dns-nextdns-webhook" \
      org.opencontainers.image.url="https://github.com/brethubbard/external-dns-nextdns-webhook" \
      org.opencontainers.image.source="https://github.com/brethubbard/external-dns-nextdns-webhook" 