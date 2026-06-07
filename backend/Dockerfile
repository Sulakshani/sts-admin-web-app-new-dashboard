# Stage 1: Build
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Stage 2: Run
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Create Choreo-required non-root user
RUN addgroup -g 10014 choreo && \
    adduser -D -u 10014 -G choreo choreouser

COPY --from=builder /app/server /app/server

ENV PORT=8080
USER 10014
EXPOSE 8080

ENTRYPOINT ["/app/server"]