FROM alpine:latest

COPY vrctlviz-alpine /vrctlviz
RUN chmod +x /vrctlviz \
&& apk add --no-cache iproute2

ENTRYPOINT [ "/vrctlviz", "run" ]

