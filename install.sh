#!/bin/bash

#run app on login
mkdir -p $HOME/.config/upstart

touch $HOME/.config/upstart/desktopOpen.conf

cat <<EOT >> $HOME/.config/upstart/desktopOpen.conf
description "Desktop Open Task"
start on desktop-start
task
script
day_watch --login
nohup x-monitor.sh &
end script
EOT

#run app on logout
touch $HOME/.config/upstart/desktopClose.conf

cat <<EOT >> $HOME/.config/upstart/desktopClose.conf
description "Desktop Close Task"
start on session-end
task
script
day_watch --logout
end script
EOT

cp ./x-monitor.sh $GOPATH/bin
