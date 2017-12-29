FROM alpine:latest

COPY rootfs /

RUN chmod +x /vrctlviz \
chmod +x /run.sh \
&& apk add --no-cache iproute2

ENTRYPOINT [ "/run.sh" ]

