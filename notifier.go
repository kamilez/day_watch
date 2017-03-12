package main

import (
	"github.com/mqu/go-notify"
	"os/exec"
)

const IMAGE_PATH = "/home/kamil/Documents/busy_beaver.png"

type Notifier interface {
	Notify() error
}

type GnomeNotification struct {
	Noti        *notify.NotifyNotification
	Title       string
	Image       string
	Information []string
}

func NewGnomeNotification(image string, title string, info ...string) Notifier {

	var information string
	for _, v := range info {
		information += v
	}

	notify.Init("daywatch")
	noti := notify.NotificationNew(title, information, IMAGE_PATH)

	return GnomeNotification{Title: title, Information: info, Noti: noti, Image: IMAGE_PATH}
}

func (n GnomeNotification) Notify() error {

	return n.Noti.Show()
}

type Zenity struct {
	command     string
	title       string
	image       string
	information []string
}

func NewZenity(title string, image string, info ...string) Notifier {
	return Zenity{"zenity", title, image, info}
}

func (z Zenity) Notify() error {

	var img string
	if z.image != "" {
		img = z.image
	} else {
		img = IMAGE_PATH
	}

	_, err := exec.Command(z.command, "--info", "--text",
		z.title, "--window-icon", img).Output()
	return err
}
