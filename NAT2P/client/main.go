package main

import "github.com/sirupsen/logrus"

const reply2peer = "ok"

var clientID = []byte("ABCD")
var svcAddr = "1.14.102.100:9998"
var localAddr = "0.0.0.0:9999"

func main() {

	peers, err := tcpRegister(localAddr, svcAddr)
	if err != nil {
		logrus.Errorf("%v\n", err)
		return
	}
	logrus.Infof("TCP注册返回消息: %v | %v\n", peers, err)

	peers2, err2 := udpRegister(localAddr, svcAddr)
	if err2 != nil {
		logrus.Errorf("%v\n", err2)
		return
	}
	logrus.Infof("UDP注册返回消息: %v | %v\n", peers2, err2)

}
