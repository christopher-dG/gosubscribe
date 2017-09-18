#!/usr/bin/env bash

source /go/src/app/config.sh

if [[ $1 = 'bash' ]]; then
    exec bash
else
    exec $GOPATH/bin/discord &
    exec $GOPATH/bin/irc
fi
