package main

//RunMode represents the available run modes for the app, production vs development so far
//if you update the RunModes, you need to run
//go get -u  golang.org/x/tools/cmd/stringer
//go generate constants.go
type RunMode int

//go:generate stringer -type=RunMode
const (
	production RunMode = iota
	development
)

//mode by default is development
var mode = development
