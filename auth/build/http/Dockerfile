FROM golang:1.20.3 AS builder
LABEL maintainer="taufikjanuar35@gmail.com"

RUN go version

WORKDIR /auth
COPY ./auth/go.mod ./
COPY ./auth/go.sum ./

RUN go mod download

COPY ./auth .

# Build Go App
RUN CGO_ENABLED=0 GOOS=linux go build -o auth-service ./cmd

FROM alpine:3.11.3

WORKDIR /app

RUN mkdir cmd docs

COPY --from=builder ./auth/cmd/ ./cmd
COPY --from=builder ./auth/docs/ ./docs
COPY --from=builder ./auth/auth-service .

# Expose port
EXPOSE 8080

# Command to run the executeable
ENTRYPOINT ["./auth-service","http"]
