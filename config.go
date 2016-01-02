package main

import (
	"github.com/elastic/libbeat/logp"
	"github.com/elastic/libbeat/outputs"
	"github.com/elastic/libbeat/publisher"
)

type RouteConfig struct {
	Period         *int64
	PublishHops    *bool
	PublishSummary *bool
	PublishChanges *bool
	MaxHops        *int
	TimeoutMs      *int
	Retries        *int
	PacketSize     *int
	Targets        *[]string
}

type ConfigSettings struct {
	Input   RouteConfig
	Output  map[string]outputs.MothershipConfig
	Logging logp.Logging
	Shipper publisher.ShipperConfig
}

var Config ConfigSettings
