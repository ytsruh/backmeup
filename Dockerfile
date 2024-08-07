# Stage 1: Build the Go app
FROM golang:alpine AS builder

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Move to working directory /build
WORKDIR /build

# Copy the code into the container
COPY . .

# Build the application
RUN go build -o main .

# Stage 2 (final stage): the running container
FROM scratch

# Copy the output from the builder stage
COPY --from=builder /build/main /app/

# Command to run
ENTRYPOINT ["/app/main"]