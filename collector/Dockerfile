FROM alpine:latest

ENV ETCDCTL_ENDPOINTS=http://etcd:2379
COPY rootfs/ /

RUN apk add --no-cache util-linux \
&& chmod +x /vrctlviz \
&& chmod +x /usr/bin/collect.sh \
&& sed -i 's/\r//g' /usr/bin/collect.sh \
&& apk add --no-cache iproute2 bash

ENTRYPOINT ["/vrctlviz", "run"]

