# syntax=docker/dockerfile:1.3

# Step 1 build binary
FROM docker.io/golang:1.22.6-alpine3.20 as builder

# Set working directory
WORKDIR /build

# Cache dependencies through image layer caching
# dependencies are downloaded and verified only if go.mod or go.sum change
COPY ./go.mod ./go.sum ./
RUN go mod download
RUN go mod verify

# Copy over rest of source files and build executable
COPY . ./
RUN go build ./cmd/raiju/raiju.go

# Separate stage for final binary in order to minimize size of image
FROM docker.io/alpine:3.20

WORKDIR /

# Connect image to repository
LABEL org.opencontainers.image.source https://github.com/nyonson/raiju
LABEL org.opencontainers.image.description "Your friendly bitcoin lightning network helper"
LABEL org.opencontainers.image.licenses MIT

# Copy over the executable from the builder stage
COPY --from=builder /build/raiju . 

# Signal that this container is meant to just run the raiju executable
ENTRYPOINT ["/raiju"]

# Default to running dashboard
CMD ["/raiju"]
