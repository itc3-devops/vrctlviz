FROM scratch

COPY vrctlviz-alpine /vrctlviz

ENTRYPOINT ["/vrctlviz", "run"]