FROM ubuntu
MAINTAINER Bernard Van De Walle <bernard@aporeto.com>

RUN mkdir -p /opt/trireme

ADD collector /opt/trireme/collector

WORKDIR /opt/trireme/collector

CMD ["./collector"]
