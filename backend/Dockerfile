FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/main.go

FROM alpine:latest
RUN addgroup -g 10014 choreo && adduser -D -u 10014 -G choreo choreouser
WORKDIR /app
COPY --from=builder /app/server .
USER 10014
EXPOSE 8083
CMD ["./server"]