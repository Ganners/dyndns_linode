DynDNS for Linode implementation in Go
======================================

This has been developed specifically for Linode (sorry!). It runs the
process in a daemon.

It will basically update the IP address to your current external IP on
the specified domain/subdomains.

Installation
------------

    go get github.com/Ganners/dyndns_linode

Usage
-----

### To start (can modify start.bash)

    dyndns_linode --configFile=FILE_PATH

### To kill the process

    dyndns_linode --stop
