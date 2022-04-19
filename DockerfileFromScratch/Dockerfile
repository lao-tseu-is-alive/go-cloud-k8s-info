# Start from the latest golang base image
FROM golang:1-alpine3.15 as builder

# Add Maintainer Info
LABEL maintainer="cgil"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY *.go ./

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o go-info-server .


######## Start a new stage from scratch #######
FROM scratch
# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/go-info-server .
# Expose port 8080 to the outside world, go-info-server will use the env PORT as listening port or 8080 as default
EXPOSE 8080
# Command to run the executable
CMD ["./go-info-server"]