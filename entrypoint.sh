#!/usr/bin/env bash

source $GOPATH/src/app/config.sh

if [[ $1 = 'bash' ]]; then
    exec bash
elif [[ $1 = 'discord' ]]; then
    exec $GOPATH/bin/discord
elif [[ $1 = 'irc' ]]; then
    exec $GOPATH/bin/irc
elif [[ $1 = 'notify' ]]; then
    exec $GOPATH/bin/notify
else
    exec $GOPATH/bin/discord &
    exec $GOPATH/bin/irc
fi
