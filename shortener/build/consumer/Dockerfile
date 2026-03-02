FROM golang:1.25 AS builder

WORKDIR /shortener
COPY ./shortener/go.mod ./
COPY ./shortener/go.sum ./

RUN go mod download

COPY ./shortener .

RUN CGO_ENABLED=0 GOOS=linux go build -o shortener-service ./cmd

FROM alpine:3.11.3

WORKDIR /app

COPY --from=builder ./shortener/.env ./.env
COPY --from=builder ./shortener/shortener-service .

ENTRYPOINT ["./shortener-service","consumer"]
