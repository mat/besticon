FROM alpine:3.7
LABEL maintainer="Matthias Luedtke (matthiasluedtke)"

RUN apk add --no-cache ca-certificates
# Fixes 'Get https://github.com/: x509: failed to
# load system roots and no roots provided'

ADD bin/linux_amd64/iconserver /var/www/iconserver

EXPOSE 8080
ENV PORT=8080

ENV HOST_ONLY_DOMAINS=*
ENV POPULAR_SITES=bing.com,github.com,instagram.com,reddit.com

WORKDIR /var/www
CMD ["./iconserver"]
