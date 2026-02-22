package main

type Daemon interface {
	ID() string
	Start() error
	Stop() error
}
