routebeat
========

<a href="https://travis-ci.org/xeb/routebeat">
<img src="https://api.travis-ci.org/xeb/routebeat.svg?branch=master" />
</a>
<a href="https://codecov.io/github/xeb/routebeat?branch=master">
  <img src="https://codecov.io/github/xeb/routebeat/coverage.svg?branch=master" alt="Coverage via Codecov" />
</a>

*For constantly tracing routes*

routebeat sends ICMP pings to a list of targets to record TCP/IP routing information.  
It uses [github.com/aeden/traceroute](https://github.com/aeden/traceroute) for
sending/recieving ping packets and tracing routes.  As well as
[elastic/libbeat](https://github.com/elastic/libbeat) to talk to
Elasticsearch and other outputs.  Essentially, those two libraries do
all the heavy lifting, routebeat is just glue around them.

Routebeat has three events it can publish, including:
- Route summary stats (with a type of "route")
- Route hop messages (with a type of "route_hop")
- Route changes during a beat run (with a type of "route_change")

By default, only the first route event is published.

## Requirements

routebeat has the same requirements around the Go environment as
libbeat, see
[here](https://github.com/elastic/beats/blob/master/CONTRIBUTING.md#dependencies).

## Installation

Install and configure [Go](https://golang.org/doc/install).

Install and update this go package with `go get -u
github.com/xeb/routebeat`.  The `routebeat` binary will then be
available in `$GOPATH/bin`.

If intending on using the Elasticsearch output, you should add a
new index template using the
[supplied one](etc/routebeat-template.json), for example with `curl
-XPUT  /_template/routebeat -d @/path/to/routebeat-template.json`.

## Usage

See the [example configuration file](etc/routebeat-example.yml) for configuring
your targets and assigning an output (default output is
Elasticsearch).

Once you've created a configuration file you can run
routebeat with `routebeat -c /path/to/pingbeat.yml`.

*NOTE:* you will likely need to run `sudo routebeat` in order to send ICMP pings.
If you'd like to see everything routebeat is doing, run something like:
`sudo ./routebeat -e -v -d routebeat -c etc/pingbeat.yml` which will output information
from the Debug logger "routebeat".

## Kibana Dashboard

There is a Kibana [export](etc/routebeat-dashboard.json) you can use to
create some basic visulizations and a simple dashboard to explore
routebeat data.

<img src="http://epicapp.com/routebeat-dashboard.png" />

### Note on privileges

In order to send regular ICMP ping packets, routebeat needs to open raw
sockets, which can only be done with superuser privileges.  So you
either need to run routebeat with sudo or as root to send regular
pings.  I haven't tried a non-priviledged UDP traceroute yet.  

Feel free to submit a PR if that is useful:)

## License

pingbeat is licensed under the Apache 2.0 [license](LICENSE).
