FROM ubuntu:trusty
MAINTAINER Matthias Luedtke (matthiasluedtke)

RUN apt-get install -y -q ca-certificates
# Fixes 'Get https://github.com/: x509: failed to
# load system roots and no roots provided'

ADD bin/linux_amd64/iconserver /var/www/iconserver

EXPOSE 8080
ENV PORT=8080
WORKDIR /var/www
CMD ["./iconserver"]
