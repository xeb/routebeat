package main

import (
	"bytes"
	"fmt"
	"github.com/aeden/traceroute"
	"github.com/elastic/libbeat/beat"
	"github.com/elastic/libbeat/cfgfile"
	"github.com/elastic/libbeat/common"
	"github.com/elastic/libbeat/logp"
	"github.com/elastic/libbeat/publisher"
	"os"
	"strings"
	"time"
)

const NANO_TO_MS = 1000000

// Routebeat struct contains all options and targets to trace
type Routebeat struct {
	period time.Duration
	config ConfigSettings
	events publisher.Client
	done   chan struct{}

	// Values are set in this struct to set defaults in Config function
	publishHops    bool
	publishSummary bool
	publishChanges bool
	maxHops        int
	timeoutMs      int
	retries        int
	packetSize     int
	targets        []string

	// store the previous hop addresses to identify route changes
	prevRoutes map[string]*Route
}

// Route struct used to compare any routing changes
type Route struct {
	hops       []byte // full concat array of all IPv4 addresses
	count      int    // total count of all hops
	durationMs int64  // total duration in milliseconds
}

// Used to determine routing changes
type RouteChange struct {
	prev  *Route
	new   *Route
	event common.MapStr
}

// Config reads in the routebeat configuration file, validating
// configuration parameters and setting default values where needed
func (r *Routebeat) Config(b *beat.Beat) error {

	// Read in provided config file, bail if problem
	err := cfgfile.Read(&r.config, "")
	if err != nil {
		logp.Err("Error reading configuration file: %v", err)
		return err
	}

	// Use period provided in config or default to 60s
	if r.config.Input.Period != nil {
		r.period = time.Duration(*r.config.Input.Period) * time.Second
	} else {
		r.period = 60 * time.Second
	}
	logp.Debug("routebeat", "Period %v\n", r.period)

	// Fill the targets array
	r.targets = make([]string, len(*r.config.Input.Targets))
	if r.config.Input.Targets != nil {
		for i, target := range *r.config.Input.Targets {
			r.targets[i] = target
			logp.Debug("routebeat", "Adding target %s\n", target)
		}
	} else {
		logp.Critical("Error: no targets specified, cannot continue!")
		os.Exit(1)
	}

	// Publish each of the Hops
	if r.config.Input.PublishSummary != nil {
		r.publishSummary = *r.config.Input.PublishSummary
	} else {
		r.publishSummary = true
	}
	logp.Debug("routebeat", "Publish Summary is %d\n", r.publishSummary)

	// Publish each of the Hops
	if r.config.Input.PublishHops != nil {
		r.publishHops = *r.config.Input.PublishHops
	} else {
		r.publishHops = false
	}
	logp.Debug("routebeat", "Publish Hops is %d\n", r.publishHops)

	// Publish each of the Hops
	if r.config.Input.PublishChanges != nil {
		r.publishChanges = *r.config.Input.PublishChanges
	} else {
		r.publishChanges = false
	}
	logp.Debug("routebeat", "Publish Changes is %d\n", r.publishChanges)

	// Set maximum hops
	if r.config.Input.MaxHops != nil {
		r.maxHops = *r.config.Input.MaxHops
	} else {
		r.maxHops = 64
	}
	logp.Debug("routebeat", "MaxHops %d\n", r.maxHops)

	// Set timeout in milliseconds
	if r.config.Input.TimeoutMs != nil {
		r.timeoutMs = *r.config.Input.TimeoutMs
	} else {
		r.timeoutMs = 500
	}
	logp.Debug("routebeat", "TimeoutMs %d\n", r.timeoutMs)

	// Set retries
	if r.config.Input.Retries != nil {
		r.retries = *r.config.Input.Retries
	} else {
		r.retries = 3
	}
	logp.Debug("routebeat", "Retries %d\n", r.retries)

	// Set packet size
	if r.config.Input.PacketSize != nil {
		r.packetSize = *r.config.Input.PacketSize
	} else {
		r.packetSize = 52
	}
	logp.Debug("routebeat", "PacketSize %d\n", r.packetSize)

	return nil
}

// Initialize without the context of a Beat
// This is called within Setup
func (r *Routebeat) Init() {
	r.prevRoutes = make(map[string]*Route)
}

// Setup performs boilerplate Beats setup
func (r *Routebeat) Setup(b *beat.Beat) error {
	r.events = b.Events
	r.done = make(chan struct{})
	r.Init()
	return nil
}

// Run the main routebeat loop
func (r *Routebeat) Run(b *beat.Beat) error {
	var err error
	ticker := time.NewTicker(r.period)
	defer ticker.Stop()

	// Do one trace right away, then start ticking
	r.TraceAllTargets()

	for {
		select {
		case <-r.done:
			return nil
		case <-ticker.C:
			timerStart := time.Now()
			r.TraceAllTargets()
			timerEnd := time.Now()
			duration := timerEnd.Sub(timerStart)
			if duration.Nanoseconds() > r.period.Nanoseconds() {
				logp.Warn("Ignoring tick(s) due to processing taking longer than one period")
			}
		}
	}

	return err
}

// Cleanup anything
func (r *Routebeat) Cleanup(b *beat.Beat) error {
	return nil
}

func (r *Routebeat) Stop() {
	close(r.done)
}

func (r *Routebeat) TraceAllTargets() {
	for _, target := range r.targets {
		msgs, er := r.TraceTarget(target)
		if er != nil {
			logp.Warn(fmt.Sprintf("Error tracing route to %s, %s", target, er))
			continue
		}

		for _, msg := range msgs {
			logp.Debug("routebeat", fmt.Sprintf("Publishing message %s, %s", msg["type"], target))
			r.events.PublishEvent(msg)
		}
	}
}

func (r *Routebeat) TraceTarget(target string) (msgs []common.MapStr, er error) {
	msgs = make([]common.MapStr, 0)
	opt := &traceroute.TracerouteOptions{}
	opt.SetTimeoutMs(r.timeoutMs)
	opt.SetRetries(r.retries)
	opt.SetMaxHops(r.maxHops)
	opt.SetPacketSize(r.packetSize)

	logp.Debug("routebeat", fmt.Sprintf("Tracing route to %s", target))
	result, er := traceroute.Traceroute(target, opt)
	if er != nil {
		return
	}

	sc := 0        // success count
	ec := 0        // error count
	sd := int64(0) // total duration of success events in nanoseconds
	ed := int64(0) // total duration of error events in nanoseconds
	route := &Route{
		hops:  make([]byte, len(result.Hops)*4),
		count: len(result.Hops),
	}

	for i, e := range result.Hops {
		durationMs := e.ElapsedTime.Nanoseconds() / NANO_TO_MS

		if r.publishHops {
			add := fmt.Sprintf("%d.%d.%d.%d", e.Address[0], e.Address[1], e.Address[2], e.Address[3])
			// From: https://github.com/aeden/traceroute/blob/master/traceroute.go#L120
			hop := common.MapStr{
				"@timestamp":  common.Time(time.Now()),
				"type":        "route_hop",
				"target":      target,
				"hop_number":  i + 1,
				"success":     e.Success,
				"address":     add,
				"host":        e.Host,
				"n":           e.N,
				"duration_ms": durationMs,
				"ttl":         e.TTL,
			}
			logp.Debug("routebeat", fmt.Sprintf("Enqueued hop event %d to %s (%t) %s", i+1, target, e.Success, add))
			// r.events.PublishEvent(hop)
			msgs = append(msgs, hop)
		}

		// Record the aggregate results for each hop
		if e.Success {
			sc++
			sd = sd + e.ElapsedTime.Nanoseconds()
		} else {
			ec++
			ed = ed + e.ElapsedTime.Nanoseconds()
		}

		route.hops[4*i] = e.Address[0]
		route.hops[4*i+1] = e.Address[1]
		route.hops[4*i+2] = e.Address[2]
		route.hops[4*i+3] = e.Address[3]

		route.durationMs = route.durationMs + durationMs
	}

	if r.publishChanges {
		// Have we traced this before? and is it different
		rc := r.GetRouteChange(target, route)
		if rc != nil {
			logp.Info("routebeat", fmt.Sprintf("Enqueued route change event\n-----\nOld route:%s\nNew route:%s\n", rc.event["PrevRoute"], rc.event["NewRoute"]))
			// r.events.PublishEvent(rc.event)
			msgs = append(msgs, rc.event)
		}
	}

	// Save the previous hops to detect route changes
	r.prevRoutes[target] = route

	if r.publishSummary {
		da := result.DestinationAddress
		event := common.MapStr{
			"@timestamp":     common.Time(time.Now()),
			"type":           "route",
			"target":         target,
			"destination":    fmt.Sprintf("%d.%d.%d.%d", da[0], da[1], da[2], da[3]),
			"hop_count":      sc + ec,
			"success_count":  sc,
			"error_count":    ec,
			"success_sum_ms": sd / NANO_TO_MS,
			"error_sum_ms":   ed / NANO_TO_MS,
			"success":        ec == 0 && sc > 0,
		}

		// Calculate the average (Elastic will be able to do this aggregation in the future)
		if sc > 0 {
			event["success_avg_ms"] = sd / int64(sc) / NANO_TO_MS
		}

		if ec > 0 {
			event["error_avg_ms"] = ed / int64(ec) / NANO_TO_MS
		}

		// r.events.PublishEvent(event)
		msgs = append(msgs, event)
		logp.Debug("routebeat", fmt.Sprintf("Enqueued summary event with %d hops", len(result.Hops)))
	}
	er = nil
	return
}

func (r *Routebeat) GetRouteChange(target string, new *Route) *RouteChange {
	if r.prevRoutes[target] == nil {
		return nil
	}

	// If no change, return nil
	if bytes.Equal(r.prevRoutes[target].hops, new.hops) {
		return nil
	}

	// Something changed
	rc := &RouteChange{
		prev: r.prevRoutes[target],
		new:  new,
	}

	rc.event = common.MapStr{
		"@timestamp":         common.Time(time.Now()),
		"type":               "route_change",
		"target":             target,
		"prev_duration_ms":   rc.prev.durationMs,
		"new_duration_ms":    rc.new.durationMs,
		"prev_hop_count":     rc.prev.count,
		"new_hop_count":      rc.new.count,
		"change_duration_ms": rc.new.durationMs - rc.prev.durationMs,
		"change_hop_count":   rc.new.count - rc.prev.count,
		"prev_route":         rc.prev.String(),
		"new_route":          rc.new.String(),
	}

	return rc
}

func (r *Route) String() string {
	if r == nil {
		return ""
	}
	adds := make([]string, r.count)
	for i := 0; i < r.count; i++ {
		add := fmt.Sprintf("%d.%d.%d.%d", r.hops[i*4], r.hops[i*4+1], r.hops[i*4+2], r.hops[i*4+3])
		adds[i] = add
	}

	return strings.Join(adds, ",")
}
