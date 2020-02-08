package main

type CommandProvider interface {
	Start() error
	Stop() error
}
