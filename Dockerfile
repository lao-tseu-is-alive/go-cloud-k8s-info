# Start from the latest golang base image
FROM golang:1.22.5-alpine3.20 AS builder

ENV PATH /usr/local/go/bin:$PATH
ENV GOLANG_VERSION 1.22.5


# Add Maintainer Info
LABEL maintainer="cgil"
LABEL org.opencontainers.image.title="go-cloud-k8s-shell"
LABEL org.opencontainers.image.description="This is a go-cloud-k8s-info container image, a simple microservice written in Golang that gives some runtime information"
LABEL org.opencontainers.image.url="https://ghcr.io/lao-tseu-is-alive/go-cloud-k8s-info:latest"
LABEL org.opencontainers.image.authors="cgil"
LABEL org.opencontainers.image.licenses="MIT"

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


######## Start a new stage  #######
# using from scratch for size and security reason
# Containers Are Not VMs! Which Base Container (Docker) Images Should We Use?
# https://blog.baeke.info/2021/03/28/distroless-or-scratch-for-go-apps/
# https://github.com/vfarcic/base-container-images-demo
# https://youtu.be/82ZCJw9poxM
FROM scratch
# to comply with security best practices
# Running containers with 'root' user can lead to a container escape situation (the default with Docker...).
# It is a best practice to run containers as non-root users
# https://docs.docker.com/develop/develop-images/dockerfile_best-practices/
# https://docs.docker.com/engine/reference/builder/#user
USER 12121:12121
WORKDIR /goapp
COPY certificates/isrg-root-x1-cross-signed.pem /goapp/certificates/
# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/go-info-server .

# Expose port 8080 to the outside world, go-info-server will use the env PORT as listening port or 8080 as default
EXPOSE 8080

# Health check to ensure the app is running
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# Command to run the executable
CMD ["./go-info-server"]
