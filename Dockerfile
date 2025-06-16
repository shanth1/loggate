FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/loggate_server ./cmd/loggate

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/loggate_server /app/loggate_server

EXPOSE 10514/udp
EXPOSE 9091
CMD ["/app/loggate_server"]
