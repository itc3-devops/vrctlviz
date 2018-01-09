#!/bin/bash

ss -itu | grep -v State |  awk 'NR%2{printf "%s ",$0;next;}1' | grep -v 127.0.0.1
