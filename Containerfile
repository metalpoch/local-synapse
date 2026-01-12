FROM golang:1.25-alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o synapse ./cmd/api/main.go
RUN CGO_ENABLED=1 GOOS=linux go build -o mcp ./cmd/mcp/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/synapse .
COPY --from=builder /app/mcp .
EXPOSE 8080
CMD ["./synapse"]
