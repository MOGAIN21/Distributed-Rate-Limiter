#Multi stage build for smaller final image
FROM golang:1.24-alpine AS builder

#Install build dependencies
RUN apk add --no-cache git make protobuf-dev

WORKDIR /app

#Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .

#Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /bin/server ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /bin/server .

COPY config/config.yaml config/

EXPOSE 50051    8080

CMD ["./server"]
