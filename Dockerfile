# Build stage
FROM golang:1.25.3-alpine3.22 AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go

# Run stage
FROM alpine:3.22

# Install netcat (nc) to wait for Postgres
RUN apk add --no-cache netcat-openbsd

WORKDIR /app
COPY --from=builder /app/main .
COPY config.yaml .
COPY start.sh .

RUN chmod +x start.sh

EXPOSE 8080

# Run the script instead of main directly
CMD ["/app/start.sh"]
