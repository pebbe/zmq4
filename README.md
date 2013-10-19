
## THIS IS A WORK IN PROGRESS

A Go interface to [ZeroMQ](http://www.zeromq.org/) version 4.

This requires ZeroMQ version 4.0.1 or above.

For ZeroMQ version 3, see: http://github.com/pebbe/zmq3

For ZeroMQ version 2, see: http://github.com/pebbe/zmq2

Including all examples of [ØMQ - The Guide](http://zguide.zeromq.org/page:all).

Keywords: networks, distributed computing, message passing, fanout, pubsub, pipeline, request-reply

## Status

It builds, that's all. No guaranties yet.

### To do:

 * Rewrite Monitor() and RecvEvent()
 * Test all old, zmq3 functionality
 * Test all functionality that is new in zmq4
 * Add link on http://zeromq.org/bindings:go
 * Add link on http://code.google.com/p/go-wiki/wiki/Projects#Networking
 * Announce on golang-nuts
 * Update zmq2 and zmq3 with links to zmq4 (README and package doc)

## Install

    go get github.com/pebbe/zmq4

## Docs

 * [package help](http://godoc.org/github.com/pebbe/zmq4)
