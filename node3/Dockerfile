# Build stage
FROM golang:1.23 as build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Ensure CGO is disabled for Alpine compatibility
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go-eth-app

# Final stage
FROM alpine:latest
WORKDIR /root/
COPY --from=build /go-eth-app .
COPY .env .env
EXPOSE 3004
CMD ["./go-eth-app"]
