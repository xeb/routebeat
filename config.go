package main

import (
	"github.com/elastic/libbeat/logp"
	"github.com/elastic/libbeat/outputs"
	"github.com/elastic/libbeat/publisher"
)

// RouteConfig is a wrapper for config
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

// ConfigSettings is beat config
type ConfigSettings struct {
	Input   RouteConfig
	Output  map[string]outputs.MothershipConfig
	Logging logp.Logging
	Shipper publisher.ShipperConfig
}

// Config is beat config
var Config ConfigSettings
