FROM alpine:latest

COPY rootfs/ /

RUN chmod +x /sbin/runit-docker \
&& sed -i 's/\r//g' /sbin/runit-docker \
&& chmod +x /usr/bin/collect.sh \
&& sed -i 's/\r//g' /usr/bin/collect.sh \
&& apk add --no-cache iproute2 bash runit

ENTRYPOINT ["/sbin/runit-docker"]

