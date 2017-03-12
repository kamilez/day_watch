#!/bin/bash

dbus-monitor --session "type='signal',interface='com.ubuntu.Upstart0_6'" | \
(
	while true; do
		read X
		if echo $X | grep "desktop-lock" &> /dev/null; then
			daywatch --logout
		elif echo $X | grep "desktop-unlock" &> /dev/null; then
			daywatch --login
		fi
	done
)

