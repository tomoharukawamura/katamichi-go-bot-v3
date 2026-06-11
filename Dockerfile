FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o bot .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Tokyo
WORKDIR /app
COPY --from=builder /app/bot .
RUN mkdir -p /data
VOLUME ["/data"]
CMD ["./bot"]
