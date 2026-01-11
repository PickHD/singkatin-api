FROM golang:1.20.3 AS builder
LABEL maintainer="taufikjanuar35@gmail.com"

RUN go version

WORKDIR /shortener
COPY ./shortener/go.mod ./
COPY ./shortener/go.sum ./

RUN go mod download

COPY ./shortener .

# Build Go App
RUN CGO_ENABLED=0 GOOS=linux go build -o shortener-service ./cmd

FROM alpine:3.11.3

WORKDIR /app

RUN mkdir cmd docs

COPY --from=builder ./shortener/cmd/ ./cmd
COPY --from=builder ./shortener/docs/ ./docs
COPY --from=builder ./shortener/shortener-service .

# Expose port
EXPOSE 9091

# Command to run the executeable
ENTRYPOINT ["./shortener-service","grpc"]
