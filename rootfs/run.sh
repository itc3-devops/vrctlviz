#!/bin/sh
# Source vars file if one exist
if [ ! -f ~/.vrctlvizcfg.yaml ]; then
source ~/.vrctlvizcfg.yaml
fi

chmod +x /usr/bin/vrctlviz

/usr/bin/vrctlviz run