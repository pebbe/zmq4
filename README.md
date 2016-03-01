# zmqchan [![Build Status](https://travis-ci.org/abligh/zmqchan.svg?branch=master)](https://travis-ci.org/abligh/zmqchan) [![GoDoc](http://godoc.org/github.com/abligh/zmqchan?status.png)](http://godoc.org/github.com/abligh/zmqchan) [![GitHub release](https://img.shields.io/github/release/abligh/zmqchan.svg)](https://github.com/abligh/zmqchan/releases)
zmqchan provides an idiomatic channels interface for go to zeromq

A library to provide channels for the [Go interface](https://github.com/pebbe/zmq4) to [ZeroMQ](http://www.zeromq.org/) version 4.

Keywords: zmq, zeromq, 0mq, networks, distributed computing, message passing, fanout, pubsub, pipeline, request-reply

zmqchan provides a `zmqchan.ChannelPair` class which wraps a `zmq.Socket` into a TX and RX channel.

Currently `zmq.Socket`s are not threadsafe. These are difficulty to use in combination with golang channels as you can poll on a set of sockets, or select on a set of channels, but not both. This creates problems if you want to use conventional go idioms, e.g. using a `chan bool` for ending goroutines.

This library provides a means of wrapping a `zmq.Socket` into a `zmq.ChannelPair`, which provides an Rx and Tx channel (as well as an error channel). This is loosely based on the idea of the [go-zmq binding by vaughan0](https://github.com/vaughan0/go-zmq) but works with ZMQ 4.x.

This is currently lightly tested / experimental.

### See also

 * [zmq4](https://github.com/pebbe/zmq4) - The go interface to ZMQ4 used by zmqchan
 * [go-zmq](https://github.com/vaughan0/go-zmq) - Another go interface to ZMQ4

## Install

    go get github.com/abligh/zmqchan

## Docs

 * [package help](http://godoc.org/github.com/abligh/zmqchan)
