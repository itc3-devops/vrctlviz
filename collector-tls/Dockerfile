FROM alpine:latest

COPY rootfs/ /

RUN chmod +x /collector \
&& chmod +x /usr/bin/collect.sh \
&& sed -i 's/\r//g' /usr/bin/collect.sh \
&& apk add --no-cache iproute2 bash

ENTRYPOINT ["/collector", "run"]

