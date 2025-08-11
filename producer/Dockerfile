FROM golang:1.24-alpine AS builder

WORKDIR /app
ENV CGO_ENABLED=0

COPY go.mod .
RUN go mod download
COPY . .
RUN go build -o producer ./

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/producer /app/producer
EXPOSE 8081
CMD ["/app/producer"]


