package main

import (
	"github.com/elastic/libbeat/beat"
)

var Version = "0.0.1"
var Name = "routebeat"

func main() {
	routebeat := &Routebeat{}
	b := beat.NewBeat(Name, Version, routebeat)
	b.CommandLineSetup()
	b.LoadConfig()
	routebeat.Config(b)
	b.Run()
}
