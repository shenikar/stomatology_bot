# Build stage
FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o tgbot cmd/main.go

# Final stage
FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /root
COPY --from=builder /app/tgbot /root/tgbot
RUN chmod +x /root/tgbot
CMD [ "./tgbot" ]