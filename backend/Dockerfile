# Stage 1: Build
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy dependency files first (layer cache optimization)
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build using the package path (not file path)
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/

# Stage 2: Run
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

# Choreo requires UID 10014
RUN addgroup -g 10014 choreo && \
    adduser -D -u 10014 -G choreo choreouser

WORKDIR /app
COPY --from=builder /app/server /app/server

ENV PORT=8080

USER 10014
EXPOSE 8080

CMD ["/app/server"]