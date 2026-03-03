# singkatin-api

Microservice-based URL shortener built with **Go**, implementing **Layered Architecture** with an **NGINX API Gateway**. Includes user authentication, a dashboard for managing short links with analytics, and a file upload pipeline.

## Architecture Overview

![ARCH](https://raw.github.com/PickHD/singkatin-api/master/arch_singkatin_api.png)

### Services

| Service | Framework | Port | Description |
|---|---|---|---|
| **Auth** | Gin | `8080` | Registration, login, email verification, password reset |
| **Shortener** | Echo | `8081` | Short URL resolution & redirect, visitor tracking |
| **User** | Fiber | `8082` | Profile management, dashboard, short link CRUD |
| **Upload** | — | Consumer | Avatar upload via MinIO (RabbitMQ consumer) |

### Infrastructure

| Component | Role |
|---|---|
| **NGINX** | API Gateway — SSL termination, rate limiting, CORS, path-based routing |
| **MongoDB** | Primary database |
| **Redis** | Caching & verification code storage |
| **RabbitMQ** | Async messaging (shortener CRUD, avatar uploads) |
| **gRPC** | Inter-service communication (User ↔ Shortener) |
| **Jaeger** | Distributed tracing (OpenTelemetry) |
| **MinIO** | S3-compatible object storage for avatars |

## API Gateway

All HTTP traffic goes through NGINX on ports `:80` (→ HTTPS redirect) and `:443`.

| External Path | Internal Service | Description |
|---|---|---|
| `/{code}` | Shortener | **Root-level redirect** (shortest URL) |
| `/s/{code}` | Shortener | Redirect alias |
| `/api/v1/auth/*` | Auth | Authentication endpoints |
| `/api/v1/shortener/*` | Shortener | Shortener API |
| `/api/v1/user/*` | User | User management API |
| `/health` | NGINX | Gateway health check |

**Features:** SSL/TLS termination, rate limiting (10 req/s per IP, burst 20), centralized CORS, HTTP → HTTPS redirect.

## Main Features

1. **Register** — Email verification with branded HTML emails
2. **Login** — JWT-based authentication
3. **Forgot / Reset Password** — Email-based password recovery
4. **User Profiles** — View & edit profile, avatar upload
5. **User Dashboard** — Analytics on short link visitor counts
6. **URL Shortener** — Generate, update, delete short links
7. **Short URL Redirect** — Root-level (`/{code}`) for ultra-short URLs

## Tech Stack

| Category | Technology |
|---|---|
| Language | Go 1.25 |
| HTTP Frameworks | Gin, Echo, Fiber |
| Database | MongoDB |
| Cache | Redis |
| Message Broker | RabbitMQ |
| Inter-service | gRPC + Protobuf |
| API Gateway | NGINX (SSL, Rate Limiting) |
| Object Storage | MinIO |
| Tracing | Jaeger (OpenTelemetry) |
| Containerization | Docker & Docker Compose |

## Prerequisites

1. **Docker & Docker Compose** installed on your machine
2. Rename `example.env` to `.env` inside each service directory
3. Fill your **SMTP configuration** in the auth `.env` (uncomment the SMTP lines)
4. Generate SSL certificates for the NGINX gateway :
    ```bash
    bash nginx/certs/generate-certs.sh
    ```

## Setup

1. **Build** all services :
    ```bash
    make build
    ```

2. **Build & Run** all services in background :
    ```bash
    make run
    ```

3. **Stop** all services :
    ```bash
    make stop
    ```

4. **Stop & Remove** all services and volumes :
    ```bash
    make remove
    ```

## API Testing

Import the Postman collection to test all endpoints :

```
singkatin-api.postman_collection.json
```

**Collection variables :**
| Variable | Description |
|---|---|
| `base_url` | Gateway host (default: `localhost`) |
| `access_token` | Auto-populated after login |
| `verify_code` | Email verification code |
| `short_id` | Short link ID for update/delete |
| `short_url` | Short URL code for redirect testing |

## License

[MIT](LICENSE)
