FROM golang:1.25 AS builder

WORKDIR /app

COPY go.work go.work.sum* ./
COPY auth/go.mod auth/go.sum* ./auth/
COPY proto/go.mod proto/go.sum* ./proto/
COPY shortener/go.mod shortener/go.sum* ./shortener/
COPY upload/go.mod upload/go.sum* ./upload/
COPY user/go.mod user/go.sum* ./user/

RUN cd auth && go mod download || true
RUN cd proto && go mod download || true
RUN cd shortener && go mod download || true
RUN cd upload && go mod download || true
RUN cd user && go mod download || true

COPY . .

RUN cd shortener && CGO_ENABLED=0 GOOS=linux go build -o /app/shortener-service ./cmd

FROM alpine:3.11.3

WORKDIR /app

COPY --from=builder /app/shortener/.env ./.env
COPY --from=builder /app/shortener-service .

EXPOSE 8081

ENTRYPOINT ["./shortener-service","http"]
