FROM alpine:latest

COPY rootfs/ /

RUN apk add --no-cache bash runit \
&& chmod +x /sbin/runit-docker \
&& sed -i 's/\r//g' /sbin/runit-docker


ENV TRAFFIC_URL=sample_data.json

ENTRYPOINT ["/sbin/runit-docker"]