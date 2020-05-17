ARG GO_VERSION=1.14

FROM golang:${GO_VERSION}-alpine AS builder

RUN mkdir /user && \
    echo 'nobody:x:65534:65534:nobody:/:' > /user/passwd && \
    echo 'nobody:x:65534:' > /user/group

RUN apk add --no-cache ca-certificates git

WORKDIR /src

COPY ./ ./

RUN GO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app ./cmd/scrape/main.go

FROM alpine:latest AS final

COPY --from=builder /user/group /user/passwd /etc/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app /app

USER nobody:nobody

ENTRYPOINT ["/app"]