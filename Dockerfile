FROM scratch

COPY vrctlviz-alpine /vrctlviz
RUN chmod +x /vrctlviz

ENTRYPOINT ["/bin/sh"]