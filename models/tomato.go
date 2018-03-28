package models

import (
	"log"
	"time"
)

type SessionType int

const (
	WORK SessionType = iota
	BREAK
)

type TomatoSession struct {
	timeLeft time.Duration
	Type     SessionType
	finished chan bool
	onTick   func(time.Duration, SessionType)
}

func (s *TomatoSession) Run() <-chan bool {

	time.AfterFunc(time.Minute, s.tick)

	s.finished = make(chan bool)

	return s.finished
}

func (s *TomatoSession) tick() {
	s.timeLeft -= time.Minute

	if s.timeLeft == 0 {
		s.finished <- true
	} else {
		time.AfterFunc(time.Minute, s.tick)
	}

	s.onTick(s.timeLeft, s.Type)
}

func NewTomatoSession(type_ SessionType, onTick func(time.Duration, SessionType)) *TomatoSession {

	var duration time.Duration

	switch type_ {
	case WORK:
		duration = time.Minute * 25
	case BREAK:
		duration = time.Minute * 5
	default:
		log.Panic("Undefined tomato session type")
	}

	return &TomatoSession{timeLeft: duration, onTick: onTick, Type: type_}
}
