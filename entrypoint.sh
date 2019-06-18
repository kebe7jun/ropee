#!/bin/sh

CMD="/usr/local/bin/ropee -log-file-path - "

args="splunk-url splunk-hec-url splunk-hec-token listen-addr splunk-metrics-index splunk-metrics-sourcetype timeout debug"

for i in $args
do
    env_arg=$(echo $i | sed 'y/abcdefghijklmnopqrstuvwxyz-/ABCDEFGHIJKLMNOPQRSTUVWXYZ_/')
    anv_arg_value=$(eval "echo \"\${$env_arg}\"")
    if [ ! -z "$anv_arg_value" ]; then
        CMD=$CMD"-$i $anv_arg_value "
    fi
done

$CMD
