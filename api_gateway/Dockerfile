FROM golang:1.16.5-alpine3.14 AS builder

WORKDIR /usr/local/go/src/

ADD . /usr/local/go/src/

RUN go clean --modcache
RUN go build -mod=readonly -o app cmd/gateway/main.go

FROM alpine:3.14

COPY --from=builder /usr/local/go/src/app /

EXPOSE 9000

CMD ["/app"]