FROM alpine:latest

COPY rootfs/ /

RUN chmod +x /vrctlviz

ENTRYPOINT ["vrctlviz", "api"]