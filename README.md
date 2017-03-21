Simple golang application for working hours notification.

App stores working and break sessions in sqlite database. Displays gnome notification
with beginning of the working hours, estimated timeout and dayily worked hours.
Project contains installation file that creates files with scripts launched on Ubuntu
login/logout and lock/unlock. If necessary, every session might be created manually
by writing calling app with --login / --logout flag.

Usage example:

day_watch --login  //writes login/unlock time (beginning of the session) into the database

day_watch --logout //writes logout/lock time (end of the session) into the database

day_watch --status //prints in shell table with recorded hours from the current day

day_watch --notify //dislpays gnome notification with the beginning of the working hours
                     end of the work and time being logged on the device

day_watch --break  //treat following logout/lock as beginning of the break that is not
                     considered as a part of working time,
