
FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod .
COPY go.sum .


RUN go mod download

# Copy the Go source code into the container
COPY iternal/cmd/ ./cmd

COPY ./ ./

# Build the Go application
RUN go build -o app ./cmd

# Set the entry point for the container
CMD ["./app"]
