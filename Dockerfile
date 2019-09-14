# Adapted from
# https://cloud.google.com/run/docs/quickstarts/build-and-deploy#containerizing

# Use the offical Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.13 as builder

# Copy local code to the container image.
WORKDIR /go/src/github.com/mat/besticon
COPY . .

# Build the command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN make build_linux_amd64

# Use a Docker multi-stage build to create a lean production image.
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM alpine:3.10
RUN apk add --no-cache ca-certificates

# Copy the binary to the production image from the builder stage.
COPY --from=builder /go/src/github.com/mat/besticon/bin/linux_amd64/iconserver /iconserver

ENV HOST_ONLY_DOMAINS=*
ENV POPULAR_SITES=bing.com,github.com,instagram.com,reddit.com
ENV HTTP_CLIENT_TIMEOUT=5s
ENV HTTP_MAX_AGE_DURATION=720h

# Run the web service on container startup.
CMD ["/iconserver"]
