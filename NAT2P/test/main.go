package main

import (
	"net"

	"github.com/sirupsen/logrus"
)

func main() {

	dialer, err := net.DialUDP("udp",
		// &net.UDPAddr{IP: net.IPv4(192, 168, 110, 200), Port: 52469},
		nil,
		&net.UDPAddr{IP: net.IPv4(58, 16, 89, 223), Port: 9557},
	)
	if err != nil {
		logrus.Errorf("2listen udp server error:%v\n", err)
	}
	defer dialer.Close()

	// 发送数据
	sendData := []byte("Hello")
	_, err = dialer.Write(sendData) // 发送数据
	if err != nil {
		logrus.Errorf("发送数据失败，err: %v\n", err)
		return
	}

	logrus.Infoln("等待回收数据...")
	// 接收数据
	data := make([]byte, 512)
	n := 0
	n, _, err = dialer.ReadFromUDP(data) // 接收数据
	if err != nil {
		logrus.Errorf("接收数据失败，err:%v \n", err)
		return
	}

	logrus.Infof("收到数据 %s\n", string(data[:n]))

}
