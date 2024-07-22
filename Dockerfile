FROM golang:1.22.5-alpine3.20 as builder
WORKDIR /goshort
RUN apk update && apk upgrade --available && sync && apk add --no-cache --virtual .build-deps
COPY . .
RUN go build -ldflags="-w -s" .
FROM alpine:3.20.1
RUN apk update && apk upgrade --available && sync
COPY index.html .
COPY static .
COPY favicon.ico .
COPY --from=builder /goshort/goshort /goshort
ENTRYPOINT ["/goshort"]
