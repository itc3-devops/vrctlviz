FROM alpine:latest

COPY rootfs /vrctlviz

RUN chmod +x /vrctlviz \
chmod +x /run.sh \
&& apk add --no-cache iproute2

ENTRYPOINT [ "/run.sh" ]

