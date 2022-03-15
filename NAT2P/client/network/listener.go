package network

import (
	"net"

	reuse "github.com/libp2p/go-reuseport"
	"github.com/sirupsen/logrus"
)

const Reply2peer = "ok"

func udpListen4Peers(laddr string) (err error) {

	logrus.Infof("监听本地 UDP %v\n", laddr)
	// 建立 udp 服务器
	var listenerUDP net.PacketConn
	listenerUDP, err = reuse.ListenPacket("udp", laddr)
	if err != nil {
		logrus.Errorf("UDP监听创建失败: %v\n", err)
		return
	}
	defer listenerUDP.Close() // 使用完关闭服务

	for {
		logrus.Infoln("等待接收数据")
		// 接收数据
		var data = make([]byte, MAXSIZE)
		var addr net.Addr
		var n int

		n, addr, err = listenerUDP.ReadFrom(data[:])
		if err != nil {
			logrus.Errorf("读取数据错误:%v\n", err)
			return
		}

		logrus.Infof("addr:%v\t count:%v\t data:%v\n", addr, n, string(data[:n]))

		if string(data[:n]) != Reply2peer {
			// 发送数据
			_, err = listenerUDP.WriteTo([]byte(Reply2peer), addr)
			if err != nil {
				logrus.Errorf("reply 发送数据失败:%v\n", err)
			}
		}

	}
}

func Listern(laddr string) (err error) {
	return udpListen4Peers(laddr)
}
