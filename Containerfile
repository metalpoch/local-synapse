FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o psynapse ./cmd/api/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/psynapse .
EXPOSE 8080
CMD ["./psynapse"]
