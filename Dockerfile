FROM scratch

COPY gopath/bin/vrctlviz-alpine /vrctlviz

ENTRYPOINT ["/vrctlviz", "run"]