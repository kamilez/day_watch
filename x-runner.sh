#!/bin/bash

ps cax | grep x-monitor.sh
if [ $? -eq 0 ]; then
  echo "x-monitor already running"
else
  exec x-monitor.sh &
fi
