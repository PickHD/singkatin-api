FROM golang:1.25 AS builder

WORKDIR /user
COPY ./user/go.mod ./
COPY ./user/go.sum ./

RUN go mod download

COPY ./user .

RUN CGO_ENABLED=0 GOOS=linux go build -o user-service ./cmd

FROM alpine:3.11.3

WORKDIR /app

COPY --from=builder ./user/.env ./.env
COPY --from=builder ./user/user-service .

ENTRYPOINT ["./user-service","http"]
