package main

import (
	"encoding/json"
	"net"
	"time"

	reuse "github.com/libp2p/go-reuseport"
	"github.com/sirupsen/logrus"
)

var dialer net.Conn
var listener net.PacketConn

var clientID = []byte("ABCD")
var reply2peer = "ok"

var svcUdpAddr = "0.0.0.0:9090"
var localUdpAddr = "0.0.0.0:9999"

type peer struct {
	Addr    string `json:"addr"`
	Network string `json:"network"`
}

func udpDial2Server(laddr, raddr string) (peers map[string]peer, err error) {
	dialer, err = reuse.Dial("udp", laddr, raddr)
	if err != nil {
		logrus.Errorf("连接服务器(%s)错误:%v\n", raddr, err)
	}
	defer dialer.Close()

	// 发送数据
	logrus.Infof("发送数据给服务器 %s \n", string(clientID))
	_, err = dialer.Write(clientID) // 发送数据
	if err != nil {
		logrus.Errorf("To server 发送数据失败: %v \n", err)
		return
	}

	// 接收数据
	var data [4096]byte
	// n, _, err = dialer.ReadFromUDP(data) // 接收数据
	n, err := dialer.Read(data[:])
	if err != nil {
		logrus.Errorf("接收数据失败: %v \n", err)
		return
	}

	err = json.Unmarshal(data[:n], &peers)
	if err != nil {
		logrus.Errorf("数据转换失败: %v \n", err)
		return
	}

	return
}

func udpListen4Peer(laddr string) (err error) {

	logrus.Infof("UDP监听本地 %v\n", laddr)
	// 建立 udp 服务器
	listener, err = reuse.ListenPacket("udp", laddr)
	if err != nil {
		logrus.Errorf("UDP监听创建失败: %v\n", err)
		return
	}
	defer listener.Close() // 使用完关闭服务

	for {
		logrus.Infoln("等待接收数据")
		// 接收数据
		var data [4096]byte
		var addr net.Addr
		var n int

		n, addr, err = listener.ReadFrom(data[:])

		if err != nil {
			logrus.Errorf("读取数据错误:%v\n", err)
			return
		}

		logrus.Infof("addr:%v\t count:%v\t data:%v\n", addr, n, string(data[:n]))

		if string(data[:n]) != reply2peer {
			// 发送数据
			_, err = listener.WriteTo([]byte(reply2peer), addr)

			if err != nil {
				logrus.Errorf("reply 发送数据失败:%v\n", err)
				return
			}
		}

	}
}

func udpSendmsg2Peer(msg string, laddr, raddr string) (err error) {

	logrus.Infof("向peer发送数据 \"%s\" %s -> %s \n", msg, laddr, raddr)
	dialer, err = reuse.Dial("udp", laddr, raddr)
	if err != nil {
		logrus.Errorf("2listen udp server error:%v\n", err)
	}
	defer dialer.Close()

	// 发送数据
	_, err = dialer.Write([]byte(msg)) // 发送数据
	if err != nil {
		logrus.Errorf("To peer 发送数据失败: %v\n", err)
		return
	}

	return
}

func main() {

	//1.与服务器通信，并获得
	peers, err := udpDial2Server(localUdpAddr, svcUdpAddr)
	if err != nil {
		logrus.Errorf("%v\n", err)
		return
	}
	logrus.Infof("注册返回消息: %v | %v\n", peers, err)

	//2.监听刚才与服务器通信的本地端口
	go udpListen4Peer(localUdpAddr)
	time.Sleep(time.Second)

	//3.向所有peer发送UDP请求，打通隧道
	for name, peerAddr := range peers {
		err = udpSendmsg2Peer(name, localUdpAddr, peerAddr.Addr)
		if err != nil {
			logrus.Errorf("%v\n", err)
		}
	}

	select {}
}
