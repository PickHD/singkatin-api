FROM golang:1.25 AS builder

WORKDIR /auth
COPY ./auth/go.mod ./
COPY ./auth/go.sum ./

RUN go mod download

COPY ./auth .

RUN CGO_ENABLED=0 GOOS=linux go build -o auth-service ./cmd

FROM alpine:3.11.3

WORKDIR /app

COPY --from=builder ./auth/.env ./.env
COPY --from=builder ./auth/auth-service .

EXPOSE 8080

ENTRYPOINT ["./auth-service","http"]
