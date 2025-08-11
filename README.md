# Restaurant Order System (Monorepo)

This repo contains a Go `consumer` service (Gin) and MongoDB, orchestrated via Docker Compose.

## Prerequisites

- Docker + Docker Compose

## Services

- mongo: MongoDB 7
- consumer: Go service (production build)
- consumer-dev: Go service with hot reload (Air)
- zookeeper: Zookeeper (for Kafka)
- kafka: Apache Kafka 3.6 (PLAINTEXT)

## Docker Compose commands

- Start production (builds and runs `consumer` + `mongo`):

```bash
  docker compose up --build
```

- Start development with hot reload (bind mount + Air):

```bash
  docker compose --profile dev up --build consumer-dev
```

- Stop all services:

```bash
  docker compose down
  docker compose --profile dev down
```

- Stop and remove volumes (fresh Mongo, reseed on next start):

```bash
  docker compose down -v
```

- View logs (follow):

```bash
  docker compose logs -f
```

- Kafka: list topics

```bash
  docker compose exec kafka kafka-topics.sh --list --bootstrap-server kafka:9094
```

- Kafka: create order

```bash
  docker compose exec kafka kafka-console-producer.sh --bootstrap-server kafka:9092 --topic orders
  # Copy this to terminal and close it to commit changes
  {"restaurantId":"689904ceab76a67dea61142a","items":["689904ceab76a67dea61142d"]}
```

- Connect to mongodb

```bash
  docker compose exec mongo mongosh
  # then inside shell:
  use restaurantdb
```

## APIs

### Producer (http://localhost:8081)
- Orders
  - GET `/orders` (headers: `x-org`) → list orders for restaurant
  - GET `/orders/recent` (headers: `x-org`) → recent orders (15m window, cached)
  - POST `/orders` (headers: `x-org`, body): creates order event in Kafka
    - Body:
      ```json
      { "items": [ { "id": "<itemId>", "quantity": 1 } ] }
      ```
- Restaurants
  - GET `/restaurants` → list restaurants with items
- Analytics
  - GET `/analytics/daily-aggregates?from=MM/DD/YYYY&to=MM/DD/YYYY` (headers: `x-org`) → totals per day
  - GET `/analytics/popular-items?from=MM/DD/YYYY&to=MM/DD/YYYY` (headers: `x-org`) → top items (by quantity) with revenue

### Consumer (http://localhost:8080)
- Orders
  - POST `/orders` (headers: `x-org`, body): creates order directly in DB
    - Body:
      ```json
      { "items": [ { "id": "<itemId>", "quantity": 1 } ] }
      ```

## Seeding

On first run (or after removing volumes), restaurants and items are seeded automatically. To reseed, run `docker compose down -v` and start again.

## Environment

Defaults are set in compose:

- `PORT=8080`
- `MONGODB_URI=mongodb://mongo:27017`
- `MONGODB_DATABASE=restaurantdb`
- `KAFKA_BROKER=kafka:9092`
- `REDIS_ADDR=redis:6379`
