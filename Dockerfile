# Dockerfile
# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o server ./cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o import_football_data ./scripts/import_football_data.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl
WORKDIR /root/

COPY --from=builder /app/server .
COPY --from=builder /app/import_football_data .

# Download football dataset from Kaggle
RUN curl -L -o football_data.zip -f \
    "https://www.kaggle.com/api/v1/datasets/download/martj42/international-football-results-from-1872-to-2017"

EXPOSE 8080
CMD ["./server"]
