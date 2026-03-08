# Dockerfile

# Frontend build stage
FROM node:alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci --ignore-scripts
COPY frontend/ ./
RUN npm run build

# Go build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
RUN go install github.com/swaggo/swag/v2/cmd/swag@v2.0.0-rc5

COPY . .
RUN swag init -g cmd/server/main.go -o docs/dist --parseInternal -ot json
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o import_football_data ./scripts/import_football_data.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl
WORKDIR /root/

COPY --from=builder /app/server .
COPY --from=builder /app/import_football_data .
COPY --from=builder /app/docs/dist ./docs/dist
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

# Download football dataset from Kaggle
RUN curl -L -o football_data.zip -f \
    "https://www.kaggle.com/api/v1/datasets/download/martj42/international-football-results-from-1872-to-2017"

EXPOSE 8080
CMD ["./server"]
