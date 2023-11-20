#!/bin/sh
set -e
if [ "$@" == "default-command" ];then
    if [ "$ROLE_BRIDGE" == 1 ];then
        if [ -f  /data/bridge.jsonnet ]; then
            /opt/reverse/reverse bridge -c /data/bridge.jsonnet
        else
            /opt/reverse/reverse bridge -c /opt/reverse/bridge.jsonnet
        fi
    else
        if [ -f  /data/portal.jsonnet ]; then
            /opt/reverse/reverse portal -c /data/portal.jsonnet
        else
            /opt/reverse/reverse portal -c /opt/reverse/portal.jsonnet
        fi
    fi
else
    exec "$@"
fi