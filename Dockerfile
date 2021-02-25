FROM golang:1.14.2 as builder

ARG CGO_ENABLED=0

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build

FROM alpine
COPY --from=mwader/static-ffmpeg:4.3.2 /ffmpeg /usr/local/bin/

RUN apk update && \
    apk add --no-cache tzdata

COPY --from=builder /app/bin/restream /restream
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY example.env /.env

ENTRYPOINT ["/restream"]
