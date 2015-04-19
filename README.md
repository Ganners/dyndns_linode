DynDNS implementation in Go
===========================

This has been developed specifically for Linode (sorry!). It runs the process in
a daemon in the background. You'll just want to fill in the launch-daemon.sh
script and run it. Might make sense to start it up in your Init.d.

It will basically update the IP address to your current external IP on a given
domain and sub domain.

Installation
------------

    go get github.com/Ganners/dyndns_linode

Usage
-----

### To start

    ./launch-daemon.sh

or

    dyndns_linode \
      --apikey=API_KEY \
      --domain=DOMAIN \
      --subdomain=SUBDOMAIN

### To kill the process

    dyndns_linode --stop
