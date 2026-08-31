FROM golang:alpine AS builder

WORKDIR /app

ENV GOTOOLCHAIN=auto

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o hooklet ./cmd/hooklet

FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

RUN mkdir -p /data

COPY --from=builder /app/hooklet /app/hooklet

ENV PORT=8080 \
    HOOKLET_DB=/data/hooklet.db \
    HOOKLET_TARGET=http://host.docker.internal:8000

EXPOSE 8080

VOLUME ["/data"]

ENTRYPOINT ["/app/hooklet"]
