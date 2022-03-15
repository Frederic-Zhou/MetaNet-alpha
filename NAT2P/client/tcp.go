package main

import (
	"bufio"
	"encoding/json"
	"net"
	"time"

	reuse "github.com/libp2p/go-reuseport"
	"github.com/sirupsen/logrus"
)

func tcpRegister(laddr, raddr string) (peers map[string]map[string]string, err error) {
	var dialer net.Conn
	dialer, err = reuse.Dial("tcp", laddr, raddr)
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

func tcpListen4Peers(laddr string) (err error) {

	logrus.Infof("TCP监听本地 %v\n", laddr)
	// 建立 udp 服务器
	var listenerTCP net.Listener
	listenerTCP, err = reuse.Listen("tcp", laddr)
	if err != nil {
		logrus.Errorf("TCP监听创建失败: %v\n", err)
		return
	}
	defer listenerTCP.Close() // 使用完关闭服务

	for {
		logrus.Infoln("等待接收数据")
		// 接收数据
		connTCP, e := listenerTCP.Accept()

		if e != nil {
			logrus.Errorf("读取数据错误:%v\n", err)
			continue
		}
		go func(conn net.Conn) {
			// 处理完关闭连接
			defer conn.Close()

			// 针对当前连接做发送和接受操作
			for {
				reader := bufio.NewReader(conn)
				var buf [4096]byte
				n, err := reader.Read(buf[:])
				if err != nil {
					logrus.Errorf("read from conn failed, err:%v\n", err)
					break
				}

				id := string(buf[:n])
				logrus.Infof("收到的数据：%v\n", id)

				// 将接受到的数据返回给客户端
				_, err = conn.Write([]byte(reply2peer))
				if err != nil {
					logrus.Errorf("write from conn failed, err:%v\n", err)
					break
				}
			}
		}(connTCP)

	}
}

func tcpSendData2Peer(data []byte, laddr, raddr string) (err error) {

	logrus.Infof("向peer发送数据 \"%d 字节\" %s -> %s \n", len(data), laddr, raddr)
	var dialer net.Conn
	dialer, err = reuse.Dial("tcp", laddr, raddr)
	if err != nil {
		logrus.Errorf("2listen tcp server error:%v\n", err)
		return
	}
	defer dialer.Close()

	// 发送数据
	_, err = dialer.Write(data) // 发送数据
	if err != nil {
		logrus.Errorf("To peer 发送数据失败: %v\n", err)
		return
	}

	result := [512]byte{}
	n, err := dialer.Read(result[:])
	if err != nil {
		logrus.Errorf("To peer 接收数据失败: %v\n", err)
		return
	}

	logrus.Warn("收到回复: ", string(result[:n]))

	return
}

func tcpRun() {

	//1.与服务器通信，并获得
	logrus.Infoln("开始TCP")
	peers, err := tcpRegister(localAddr, svcAddr)
	if err != nil {
		logrus.Errorf("%v\n", err)
		return
	}
	logrus.Infof("注册返回消息: %v | %v\n", peers, err)

	//2.监听刚才与服务器通信的本地端口
	go tcpListen4Peers(localAddr)
	time.Sleep(time.Second)

	//3.向所有peer发送UDP请求，打通隧道
	for name, peerAddr := range peers {
		err = tcpSendData2Peer([]byte(name), localAddr, peerAddr["tcp"])
		if err != nil {
			logrus.Errorf("%v\n", err)
		}
	}

	select {}
}
