#!/bin/bash

dbus-monitor --session "type='signal',interface='com.ubuntu.Upstart0_6'" | \
(
	while true; do
		read X
		if echo $X | grep "desktop-lock" &> /dev/null; then
			day_watch --logout
		elif echo $X | grep "desktop-unlock" &> /dev/null; then
			day_watch --login
		fi
	done
)

