FROM golang:1.25 AS builder

WORKDIR /upload
COPY ./upload/go.mod ./
COPY ./upload/go.sum ./

RUN go mod download

COPY ./upload .

RUN CGO_ENABLED=0 GOOS=linux go build -o upload-service ./cmd

FROM alpine:3.11.3

WORKDIR /app

COPY --from=builder ./upload/.env ./.env
COPY --from=builder ./upload/upload-service .

ENTRYPOINT ["./upload-service","consumer"]
