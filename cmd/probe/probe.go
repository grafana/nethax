package main

import (
	"errors"
)

type Probe interface {
	Run() error
}

var (
	errConnectionSucceeded = errors.New("connection succeeded when expecting a failure")
	errConnectionFailed    = errors.New("connection failed")
	errAssertionFailed     = errors.New("assertion failed")
)
