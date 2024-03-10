# Use the official Golang image as a base
FROM golang:1.17-alpine AS builder

# Set the current working directory inside the container
WORKDIR /app

# Copy the Go application source code to the container
COPY . .

# Install PostgreSQL client library
RUN apk --no-cache add postgresql-client

# Build the Go application
RUN go build -o app .

# Start a new stage from scratch
FROM alpine:latest

# Set the current working directory inside the container
WORKDIR /root/

# Copy the built executable from the previous stage
COPY --from=builder /app/app .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./app"]
