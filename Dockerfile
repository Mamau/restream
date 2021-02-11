FROM golang as builder

ARG CGO_ENABLED=0

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build

FROM ubuntu
RUN apt-get update && apt-get install -y ffmpeg

COPY --from=builder /app/bin/restream /home/app/restream
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/home/app/restream"]
