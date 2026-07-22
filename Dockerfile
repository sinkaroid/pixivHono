# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server main.go

FROM alpine:latest AS runtime
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=build /app/server ./server
EXPOSE 3000
CMD ["./server"]
