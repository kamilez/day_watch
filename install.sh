#!/bin/bash

cp ./x-runner.sh $GOPATH/bin
cp ./x-monitor.sh $GOPATH/bin
cp ./day_watch.sh.desktop $HOME/.config/autostart
cp ./busy_beaver.png $HOME/Documents

exec x-runner.sh
