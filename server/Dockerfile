FROM alpine:latest
ENV ETCDCTL_ENDPOINTS=http://example-etcd-cluster.vrctlviz:2379

COPY rootfs/ /

RUN chmod +x /vrctlviz

ENTRYPOINT ["vrctlviz", "api"]