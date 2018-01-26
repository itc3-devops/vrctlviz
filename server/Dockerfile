FROM alpine:latest
ENV ETCDCTL_ENDPOINTS=http://etcd:2379
ENV ETCDCTL_API=3

COPY rootfs/ /

RUN chmod +x /vrctlviz

ENTRYPOINT ["/vrctlviz", "api"]