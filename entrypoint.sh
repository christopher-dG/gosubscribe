#!/usr/bin/env bash

source $GOPATH/src/app/config.sh

if [[ $1 = 'bash' ]]; then
    exec bash
else
    exec $GOPATH/bin/discord &
    exec $GOPATH/bin/irc
fi
