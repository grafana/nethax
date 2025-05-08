package main

import (
	"context"
	"errors"
)

type Probe interface {
	Run(context.Context) error
}

var (
	errConnectionSucceeded = errors.New("connection succeeded when expecting a failure")
	errConnectionFailed    = errors.New("connection failed")
	errAssertionFailed     = errors.New("assertion failed")
)
