# Production-Ready Order Management System

Beginner-friendly Go case study implementing REST, GraphQL, PostgreSQL, Gin, JWT gateway, Consul registration, validation, structured logging, metrics, health checks, resilience, Docker, CI/CD and AWS deployment guidance.

## Services

This project has three clearly separated applications:

```text
order-management/
├── order-service/      # Main Order Management API
├── product-service/    # External Product Service
└── gateway/            # JWT Gateway
```

### Order Service
Runs on port `8080` and contains the Controller → Service → Repository flow, PostgreSQL access, GraphQL, middleware and the Product Client.

### Product Service
Runs on port `8081` and provides product name, price and stock through `GET /products/:id`, atomically reserves stock through `POST /products/:id/reserve`, and restores stock through `POST /products/release`.

### Gateway
Runs on port `8088`, creates demo JWTs through `/login`, validates Bearer tokens and forwards order/GraphQL requests to the Order Service.

## Run locally

1. Copy `.env.example` to `.env` and set values.
2. Create PostgreSQL database `orderdb`.
3. Run `db/schema.sql` (creates `products`, `product_reservations`, and `orders`).

### Existing database

If you already used an earlier version of this project, run `db/schema.sql` again. It adds `product_name`, reservation fields, the product tables, and indexes needed by the updated stock flow.

4. Start the Product Service:
   `go run ./product-service`
5. Start the Order Service:
   `go run ./order-service`
6. Start the Gateway:
   `go run ./gateway`
7. REST: `http://localhost:8080/api/v1/orders`
8. GraphiQL: `http://localhost:8080/graphql`
9. Gateway: `http://localhost:8088`

## Swagger annotations

Swagger annotations are included directly above the Order Controller methods and in `order-service/main.go`. The ZIP intentionally does not include a generated `docs/` folder or `swagger.yaml`. To generate Swagger files locally, install `swag` and run `swag init -g order-service/main.go` from the project root.

## Resilience

The Order Service calls the Product Service through `order-service/client/product_client.go`. That client demonstrates a 2-second HTTP timeout, maximum 3 attempts, exponential backoff (100ms, 200ms, 400ms), context cancellation and a circuit breaker. There are no unlimited retries.

## Consul

Run Consul locally and then execute:

`powershell -ExecutionPolicy Bypass -File .\\register-consul.ps1`

The script registers Order Service (`8080`), Product Service (`8081`) and Gateway (`8088`) with Consul.

## Tests

`go test -race ./...`

## CI/CD

GitHub Actions runs build, tests with race detector, vet, gosec and Docker build. The deploy job is a safe placeholder for AWS OIDC/ECR/EC2 deployment so no cloud credentials are embedded in the repository.

## AWS

Use EC2/ECS, ALB, RDS PostgreSQL, IAM roles, Secrets Manager and CloudWatch for production deployment.


Product service:
http://localhost:8081/health/live
http://localhost:8081/products/501

Order service:
http://localhost:8080/health/live
http://localhost:8080/health/ready
Create order - POST http://localhost:8080/api/v1/orders
Get all orders - GET http://localhost:8080/api/v1/orders
Get one order - GET http://localhost:8080/api/v1/orders/1
Update - PUT http://localhost:8080/api/v1/orders/1/status

Graphql
http://localhost:8080/graphql

query:
query {
  orders {
    id
    customerId
    productId
    quantity
    price
    status
  }
}

get order:
query {
  order(id: 1) {
    id
    customerId
    productId
    quantity
    price
    status
  }
}

create order:
mutation {
  createOrder(
    input: {
      customerId: 101
      productId: 501
      quantity: 2
    }
  ) {
    id
    customerId
    productId
    quantity
    price
    status
    createdAt
    updatedAt
  }
}

gateway auth:
http://localhost:8088/login
<!-- eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoidXNlciJ9.p7YWpOOFNGoHbPG4QfDmKhSmBp8P_50sWhNZ0wUSSaA -->
Get: http://localhost:8088/api/v1/orders

Test GraphQL through JWT Gateway
POST http://localhost:8088/graphql
Body:
{
  "query": "query { orders { id customerId productId quantity price status createdAt updatedAt } }"
}

consul:
consul agent -dev
http://localhost:8500

powershell -ExecutionPolicy Bypass -File .\register-consul.ps1