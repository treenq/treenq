package cdk

import (
	"errors"
	"sync"
)

type InMemLocker struct {
	locked bool
	mx     sync.RWMutex
}

var (
	ErrAlreadyLocked = errors.New("already locked")
)

func NewInMemLocker() *InMemLocker {
	return &InMemLocker{}
}

func (l *InMemLocker) Lock() error {
	l.mx.Lock()
	defer l.mx.Unlock()

	if l.locked {
		return ErrAlreadyLocked
	}

	l.locked = true
	return nil
}

func (l *InMemLocker) Unlock() {
	l.mx.Lock()
	l.locked = false
	l.mx.Unlock()
}
