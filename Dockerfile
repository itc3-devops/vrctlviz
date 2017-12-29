FROM scratch

COPY gopath/bin/vrctlviz /vrctlviz

ENTRYPOINT ["/vrctlviz"]