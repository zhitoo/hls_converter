FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o hls_converter .

FROM alpine:3.21

RUN apk add --no-cache ffmpeg && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /app/hls_converter .

RUN mkdir -p storage/users storage/logs tasks && \
    chown -R app:app /app

USER app

EXPOSE 8080

CMD ["./hls_converter"]
