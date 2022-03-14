package main

import (
	"net"
)

const reply2peer = "ok"

var dialer net.Conn
var listenerUDP net.PacketConn
var listenerTCP net.Listener
var clientID = []byte("ABCD")
var svcAddr = "1.14.102.100:9998"
var localAddr = "0.0.0.0:9999"

func main() {
	tcpRun()
	udpRun()
}
