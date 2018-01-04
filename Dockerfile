FROM alpine:latest

COPY rootfs/ /

RUN chmod +x /usr/bin/vrctlviz \
chmod +x /run.sh \
&& apk add --no-cache iproute2

ENTRYPOINT [ "/run.sh" ]

