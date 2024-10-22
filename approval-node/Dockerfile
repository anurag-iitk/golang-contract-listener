# Build stage
FROM golang:1.23 as build
WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go-eth-app

# Final stage
FROM alpine:latest
WORKDIR /root/

# Copy the binary from the build stage
COPY --from=build /go-eth-app .

# Expose port 4001
EXPOSE 4001

# Run the application
CMD ["./go-eth-app"]
