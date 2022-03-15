package main

const reply2peer = "ok"

var clientID = []byte("ABCD")
var svcAddr = "0.0.0.0:9998"
var localAddr = "0.0.0.0:9999"

func main() {
	tcpRun()
	udpRun()
}
