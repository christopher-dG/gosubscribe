#!/usr/bin/env bash

source $GOPATH/src/app/config.sh

case $1 in
'bash' ) exec bash ;;
'discord' ) exec $GOPATH/bin/discord ;;
'irc' ) exec $GOPATH/bin/irc ;;
'notify' ) exec $GOPATH/bin/notify ;;
* )
    exec $GOPATH/bin/discord &
    exec $GOPATH/bin/irc
    ;;
esac
