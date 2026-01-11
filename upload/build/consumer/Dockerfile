FROM golang:1.20.3 AS builder
LABEL maintainer="taufikjanuar35@gmail.com"

RUN go version

WORKDIR /upload
COPY ./upload/go.mod ./
COPY ./upload/go.sum ./

RUN go mod download

COPY ./upload .

# Build Go App
RUN CGO_ENABLED=0 GOOS=linux go build -o upload-service ./cmd

FROM alpine:3.11.3

WORKDIR /app

RUN mkdir cmd

COPY --from=builder ./upload/cmd/ ./cmd
COPY --from=builder ./upload/upload-service .

# Command to run the executeable
ENTRYPOINT ["./upload-service","consumer"]
