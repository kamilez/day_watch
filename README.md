Simple golang application for working hours notification and pomodoro timing.

App stores work and break sessions in sqlite database. Displays gnome notification
with the start time of the working hours, estimated timeout and daily worked hours.
Project contains installation file that creates files with scripts launched on Ubuntu
login/logout and lock/unlock. If necessary, every session might be created manually
by calling app with --login / --logout flag.

Prerequisites:

notify-send
zenity

Usage example:

day_watch          //runs pomodoro timer

day_watch --login  //writes login/unlock time (session start) into the database

day_watch --logout //writes logout/lock time (session end) into the database

day_watch --status //prints in shell table with recorded hours from the current day

day_watch --notify //dislpays gnome notification with the beginning of the working hours
                     end of the work and time being logged on the device

day_watch --break  //treat following logout/lock as beginning of the break that is not
                     considered as a part of work time,
