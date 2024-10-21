# Use an official Golang image to build the app
FROM golang:1.19 as builder

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files
COPY go.mod go.sum ./

# Download the dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o subscriber

# Use a smaller image to run the app
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/subscriber .

# Expose the port on which the app will run
EXPOSE 4000

# Command to run the executable
CMD ["./subscriber"]
