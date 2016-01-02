package main

import (
	"github.com/elastic/libbeat/beat"
)

// Version is the beat version
var Version = "0.0.1"

// Name is the beat name
var Name = "routebeat"

func main() {
	routebeat := &Routebeat{}
	b := beat.NewBeat(Name, Version, routebeat)
	b.CommandLineSetup()
	b.LoadConfig()
	routebeat.Config(b)
	b.Run()
}
