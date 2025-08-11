FROM golang:1.24-alpine AS builder

WORKDIR /app

ENV CGO_ENABLED=0

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o consumer ./

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/consumer /app/consumer
EXPOSE 8080
CMD ["/app/consumer"]


