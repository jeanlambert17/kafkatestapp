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
  docker compose exec kafka kafka-topics.sh --list --bootstrap-server kafka:29092
```

- Rebuild only the Go service:

```bash
  docker compose build consumer
```

- Exec into dev container shell:

```bash
  docker compose exec consumer-dev sh
```

- Connect to mongodb

```bash
  docker compose exec mongo mongosh
  # then inside shell:
  use restaurantdb
```

## Endpoints

- Consumer API: `http://localhost:8080`
  - GET `/orders`
  - GET `/orders/:id`

## Seeding

On first run (or after removing volumes), restaurants and items are seeded automatically. To reseed, run `docker compose down -v` and start again.

## Environment

Defaults are set in compose:

- `PORT=8080`
- `MONGODB_URI=mongodb://mongo:27017`
- `MONGODB_DATABASE=restaurantdb`
