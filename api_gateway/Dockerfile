FROM golang:1.22.3-alpine3.19 AS builder

WORKDIR /usr/local/go/src/

ADD . .

RUN go clean --modcache
RUN go build -mod=readonly -o app cmd/gateway/main.go

FROM alpine:3.19

COPY --from=builder /usr/local/go/src/app /

EXPOSE 9000

CMD ["/app"]