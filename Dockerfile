FROM golang:alpine AS builder
WORKDIR /app
COPY . .
RUN go build -ldflags="-s -w" -o ai-context .

FROM alpine:latest
WORKDIR /app
RUN mkdir -p /app/context
COPY --from=builder /app/ai-context .
EXPOSE 8080
CMD ["/app/ai-context"]
