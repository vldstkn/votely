FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o votely cmd/votely/main.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/votely .
COPY --from=builder /app/configs ./configs

CMD ["./votely"]