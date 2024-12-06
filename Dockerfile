# Adapted from
# https://cloud.google.com/run/docs/quickstarts/build-and-deploy#containerizing

# Use the official Golang image to create a build artifact.
# https://hub.docker.com/_/golang
FROM golang:1.23 as builder

LABEL org.opencontainers.image.source="https://github.com/mat/besticon"
LABEL org.opencontainers.image.licenses="MIT"

# Copy local code to the container image.
WORKDIR /app
COPY . .

# TARGETARCH is set only by the docker buildx command - or manually
ARG TARGETARCH

# Build the command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN make build_linux_${TARGETARCH}

# Use a Docker multi-stage build to create a lean production image.
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM alpine:3.20


# Have to define TARGETARCH again for the second stage
ARG TARGETARCH

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/bin/linux_${TARGETARCH}/iconserver /iconserver

ENV ADDRESS=''
ENV CACHE_SIZE_MB=32
ENV CORS_ENABLED=false
ENV CORS_ALLOWED_HEADERS=''
ENV CORS_ALLOWED_METHODS=''
ENV CORS_ALLOWED_ORIGINS=''
ENV CORS_ALLOW_CREDENTIALS=''
ENV CORS_DEBUG=''
ENV HOST_ONLY_DOMAINS=*
ENV HTTP_CLIENT_TIMEOUT=5s
ENV HTTP_MAX_AGE_DURATION=720h
ENV HTTP_USER_AGENT=''
ENV POPULAR_SITES=bing.com,github.com,instagram.com,reddit.com
ENV PORT=8080
ENV SERVER_MODE=redirect

ARG VERSION=''
ARG REVISION=''

LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.revision="${REVISION}"

# Run the web service on container startup.
CMD ["/iconserver"]
