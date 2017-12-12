package utils

import (
	"log"
	"os"
	"os/exec"
	"os/user"
	"strings"

	notify "github.com/mqu/go-notify"
)

var DefaultImagePath string

func init() {
	usr, err := user.Current()
	if err != nil {
		log.Fatalln(err.Error())
	}

	DefaultImagePath = usr.HomeDir + "/Documents/busy_beaver.png"
}

type Notifier interface {
	Notify(...string) error
}

type GnomeNotification struct {
	Title string
	Image string
}

func NewGnomeNotification(image string, title string) Notifier {

	notify.Init("daywatch")

	if image == "" {
		image = DefaultImagePath
	} else if _, err := os.Stat(image); !os.IsExist(err) {
		image = DefaultImagePath
	}

	return GnomeNotification{Title: title, Image: image}
}

func (n GnomeNotification) Notify(info ...string) error {

	return n.notify(strings.Join(info, "\n"))
}

func (n GnomeNotification) notify(info string) error {

	command := exec.Command("notify-send", "-i", n.Image, n.Title, info)
	return command.Run()
}
