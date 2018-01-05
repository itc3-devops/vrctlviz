FROM alpine:latest

COPY rootfs/ /

RUN apk add --no-cache bash runit \
&& chmod +x /sbin/runit-docker \
&& sed -i 's/\r//g' /sbin/runit-docker

EXPOSE 8080

ENV TRAFFIC_URL=sample_data.json
ENV TRAFFIC_INTERVAL=300

ENTRYPOINT ["/sbin/runit-docker"]