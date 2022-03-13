package main

import (
	"encoding/json"
	"net"

	"github.com/sirupsen/logrus"
)

var dialer, listener *net.UDPConn
var clientID = "ABCD"
var reply2peer = "ok"

var svcUdpAddr = net.UDPAddr{
	IP:   net.IPv4(1, 14, 102, 100),
	Port: 9090,
}

//
type peer struct {
	Addr    string `json:"addr"`
	Network string `json:"network"`
}

type registerResult struct {
	LocalAddr net.Addr
	Peers     map[string]peer `json:"peers"`
}

func udpDial2Server(raddr *net.UDPAddr) (rr registerResult, err error) {
	dialer, err = net.DialUDP("udp", nil, raddr)
	if err != nil {
		logrus.Errorf("连接%s服务器(%s)错误:%v\n", raddr.Network(), raddr.String(), err)
	}
	defer dialer.Close()

	// 发送数据
	sendData := []byte(clientID)
	logrus.Infof("发送数据给服务器 %s \n", clientID)
	_, err = dialer.Write(sendData) // 发送数据
	if err != nil {
		logrus.Errorf("发送数据失败，err:%v \n", err)
		return
	}

	// 接收数据
	data := make([]byte, 4096)
	n := 0
	n, _, err = dialer.ReadFromUDP(data) // 接收数据
	if err != nil {
		logrus.Errorf("接收数据失败，err: %v \n", err)
		return
	}

	err = json.Unmarshal(data[:n], &rr.Peers)
	if err != nil {
		logrus.Errorf("数据转换失败，err: %v \n", err)
		return
	}

	rr.LocalAddr, err = net.ResolveUDPAddr(dialer.LocalAddr().Network(), dialer.LocalAddr().String())

	return
}

func udpListen4Peer(laddr *net.UDPAddr) (err error) {

	logrus.Infof("UDP监听本地 %v\n", laddr.String())
	// 建立 udp 服务器
	listener, err = net.ListenUDP("udp", laddr)
	if err != nil {
		logrus.Errorf("UDP监听创建失败:%v\n", err)
		return
	}
	defer listener.Close() // 使用完关闭服务

	for {
		logrus.Infoln("等待接收数据")
		// 接收数据
		var data [4096]byte
		var addr *net.UDPAddr
		var n int
		n, addr, err = listener.ReadFromUDP(data[:])
		if err != nil {
			logrus.Errorf("读取数据错误:%v\n", err)
			return
		}

		logrus.Infof("addr:%v\t count:%v\t data:%v\n", addr, n, string(data[:n]))

		// 发送数据
		_, err = listener.WriteToUDP([]byte(reply2peer), addr)
		if err != nil {
			logrus.Errorf("发送数据失败:%v\n", err)
			return
		}
	}
}

func udpSendmsg2Peer(msg string, laddr, raddr *net.UDPAddr) (err error) {

	logrus.Infof("向peer发送数据 \"%s\" %s -> %s \n", msg, laddr.String(), raddr.String())
	dialer, err = net.DialUDP("udp", laddr, raddr)
	if err != nil {
		logrus.Errorf("2listen udp server error:%v\n", err)
	}
	defer dialer.Close()

	// 发送数据
	sendData := []byte(msg)
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

	if string(data[:n]) == reply2peer {
		logrus.Infoln("消息发送成功")
	} else {
		logrus.Errorln("消息返回失败")
	}

	logrus.Infof("收到数据 %s\n", string(data[:n]))

	return
}

func main() {

	//1.与服务器通信，并获得
	rr, err := udpDial2Server(&svcUdpAddr)
	if err != nil {
		logrus.Errorf("%v\n", err)
		return
	}
	logrus.Infof("注册返回消息: %v | %v\n", rr, err)

	//2.向所有peer发送UDP请求，打通隧道
	for name, addr := range rr.Peers {
		raddr, err := net.ResolveUDPAddr(addr.Network, addr.Addr)
		if err != nil {
			logrus.Errorf("%v\n", err)
			continue
		}
		err = udpSendmsg2Peer(name, rr.LocalAddr.(*net.UDPAddr), raddr)
		if err != nil {
			logrus.Errorf("%v\n", err)
		}
	}

	//3.监听刚才与服务器通信的本地端口
	udpListen4Peer(rr.LocalAddr.(*net.UDPAddr))

}
