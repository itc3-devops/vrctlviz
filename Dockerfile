FROM alpine:latest

COPY rootfs/ /

RUN chmod +x /run.sh \
&& sed -i 's/\r//g' /run.sh \
&& sed -i 's/\r//g' /usr/bin/collect.sh \
&& chmod +x /usr/bin/collect.sh \
&& apk add --no-cache iproute2

ENTRYPOINT [ "/run.sh" ]

