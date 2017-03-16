package main

import (
	"log"
)

func ErrorCheck(err error) {

	if err != nil {
		log.Fatalln(err.Error())
	}
}
