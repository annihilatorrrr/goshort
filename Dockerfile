FROM golang:1.22.5-alpine3.20 as builder
WORKDIR /goshort
RUN apk update && apk upgrade --available && sync && apk add --no-cache --virtual .build-deps
COPY . .
RUN go build -ldflags="-w -s" .
FROM alpine:3.20.2
RUN apk update && apk upgrade --available && sync
COPY --from=builder /goshort/index.html /index.html
COPY --from=builder /goshort/static /static
COPY --from=builder /goshort/favicon.ico /favicon.ico
COPY --from=builder /goshort/goshort /goshort
ENTRYPOINT ["/goshort"]
