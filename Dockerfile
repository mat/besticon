FROM alpine:3.8
LABEL maintainer="Matthias Luedtke (matthiasluedtke)"

RUN apk add --no-cache ca-certificates
# Fixes 'Get https://github.com/: x509: failed to
# load system roots and no roots provided'

# https://docs.docker.com/develop/develop-images/dockerfile_best-practices/#add-or-copy
COPY bin/linux_amd64/iconserver /var/www/iconserver

EXPOSE 8080
ENV PORT=8080

ENV HOST_ONLY_DOMAINS=*
ENV POPULAR_SITES=bing.com,github.com,instagram.com,reddit.com
ENV HTTP_CLIENT_TIMEOUT=5s

WORKDIR /var/www
CMD ["./iconserver"]
