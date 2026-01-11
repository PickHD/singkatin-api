FROM golang:1.20.3 AS builder
LABEL maintainer="taufikjanuar35@gmail.com"

RUN go version

WORKDIR /user
COPY ./user/go.mod ./
COPY ./user/go.sum ./

RUN go mod download

COPY ./user .

# Build Go App
RUN CGO_ENABLED=0 GOOS=linux go build -o user-service ./cmd

FROM alpine:3.11.3

WORKDIR /app

RUN mkdir cmd docs

COPY --from=builder ./user/cmd/ ./cmd
COPY --from=builder ./user/docs/ ./docs
COPY --from=builder ./user/user-service .

# Expose port
EXPOSE 8082

# Command to run the executeable
ENTRYPOINT ["./user-service","http"]
