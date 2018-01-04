FROM alpine:latest

COPY rootfs/ /

RUN chmod +x /sbin/runit-docker \
&& sed -i 's/\r//g' /sbin/runit-docker
&& apk add --no-cache iproute2 bash

ENTRYPOINT ["/sbin/runit-docker"]

