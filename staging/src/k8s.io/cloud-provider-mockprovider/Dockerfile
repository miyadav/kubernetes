# Step 1: Use Golang base image to build the Go application
FROM golang:1.24 as builder

# Step 2: Set the working directory inside the container
WORKDIR /workspace

# Step 3: Copy go.mod and go.sum files to the container to handle dependencies
COPY go.mod go.sum ./

# Step 4: Download Go dependencies
RUN go mod download

# Step 5: Copy the entire source code into the container
COPY . .

# Step 6: Build the Go application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cloud-provider-mock .

# Step 7: Create a minimal runtime image based on Alpine Linux
FROM alpine:3.17

# Step 8: Install necessary CA certificates (needed for secure HTTP requests)
RUN apk add --no-cache ca-certificates

# Step 9: Copy the compiled binary from the builder image
COPY --from=builder /workspace/cloud-provider-mock /usr/local/bin/cloud-provider-mock

# Step 10: Define the entry point for the application
ENTRYPOINT ["/usr/local/bin/cloud-provider-mock"]

# Step 11: (Optional) Set default arguments for running the mock provider (can be overridden)
CMD []

