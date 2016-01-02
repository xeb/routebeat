package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestTraceTarget(t *testing.T) {
	rb := &Routebeat{}
	rb.Init()
	rb.publishChanges = true
	rb.publishHops = true
	rb.publishSummary = true

	msgs, er := rb.TraceTarget("google.com")

	if er != nil {
		fmt.Printf("ERROR %s", er)
		assert.Nil(t, er)
	}

	assert.NotNil(t, msgs)
	fmt.Printf("All events are %s", rb.events)
}

func TestRouteChangeMiddleHop(t *testing.T) {
	rb := &Routebeat{}
	rb.Init()
	rb.publishChanges = true
	rb.publishHops = true
	rb.publishSummary = true

	rb.prevRoutes = make(map[string]*Route, 1)
	target := "google.com"

	prevRoute := &Route{
		count:      3,
		durationMs: 50,
		hops:       make([]byte, 3*4),
	}

	rb.prevRoutes[target] = prevRoute

	newRoute := &Route{
		count:      4,
		durationMs: 50,
		hops:       make([]byte, 4*4),
	}

	for i := 0; i < len(prevRoute.hops); i++ {
		b := byte(rand.Int31())
		prevRoute.hops[i] = b
		newRoute.hops[i] = b
	}

	rc := rb.GetRouteChange(target, newRoute)
	assert.NotNil(t, rc)

	prevRouteStr := fmt.Sprintf("%s", rc.event["prev_route"])
	newRouteStr := fmt.Sprintf("%s", rc.event["new_route"])

	diff := newRouteStr[len(prevRouteStr)+1:]

	assert.True(t, len(prevRouteStr) < len(newRouteStr), "Previous route is shorter")
	assert.Equal(t, diff, "0.0.0.0", "Last part of new route is correct")
}

func TestRouteChangeReturnsNil(t *testing.T) {
	rb := &Routebeat{}
	rb.Init()
	pv := &Route{
		hops: []byte{1, 2, 3},
	}
	rb.prevRoutes["google.com"] = pv

	nr := &Route{
		hops: []byte{1, 2, 3},
	}

	rc := rb.GetRouteChange("google.com", nr)
	assert.Nil(t, rc)
}

func TestRouteToStringIsNil(t *testing.T) {
	route := &Route{
		hops: []byte{1, 2, 3},
	}
	assert.NotNil(t, route)
	route = nil
	str := route.String()
	assert.Equal(t, str, "", "Nil route returns empty string")
}
