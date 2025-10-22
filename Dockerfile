# Build stage
FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY migrations ./migrations
RUN go build -o tgbot cmd/main.go

# Final stage
FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /root
COPY --from=builder /app/tgbot /root/tgbot
COPY --from=builder /app/migrations ./migrations
COPY credentials.json .
RUN chmod +x /root/tgbot
CMD [ "./tgbot" ]