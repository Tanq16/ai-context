# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git curl make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Download assets and build
RUN make assets && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o ai-context .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

COPY --from=builder /app/ai-context .

EXPOSE 8080
ENTRYPOINT ["./ai-context"]
CMD ["serve"]