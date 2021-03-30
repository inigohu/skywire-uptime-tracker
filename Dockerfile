FROM debian:stretch-slim

# minimal needed packages
RUN apt-get update \
&& apt-get install -y apt-transport-https ca-certificates software-properties-common

ADD ./skywire-uptime-tracker /usr/local/bin/skywire-uptime-tracker
