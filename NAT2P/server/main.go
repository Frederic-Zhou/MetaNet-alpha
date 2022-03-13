package main

import (
	"bufio"
	"encoding/json"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
)

var peersMap = sync.Map{}
var udpPort = 9090

type peer struct {
	Addr    string `json:"addr"`
	Network string `json:"network"`
}

func main() {

	go udpServer()
	go tcpServer()

	select {}
}

func udpServer() {
	// 建立 udp 服务器
	logrus.Info("启动 UDP server")
	listen, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: udpPort,
	})
	if err != nil {
		logrus.Errorf("listen failed error:%v\n", err)
		return
	}
	defer listen.Close() // 使用完关闭服务
	for {
		// 接收数据
		logrus.Info("接收UDP数据")
		var data [1024]byte
		var id string
		n, addr, err := listen.ReadFromUDP(data[:])
		if err != nil {
			logrus.Errorf("read data error:%v\n", err)
			return
		}

		//客户端发送注册的内容作为唯一识别ID，日后优化客户端的消息生成以及服务端的验证方式。
		id = string(data[:n])
		logrus.Infof("客户端地址:%v\t 数据长度:%v\t 数据:%v\n", addr, n, id)

		//保存到服务器节点列表
		peersMap.Store(id, peer{Addr: addr.String(), Network: addr.Network()})

		//将节点信息读取到一个map中
		peers := map[string]peer{}
		peersMap.Range(func(k interface{}, v interface{}) bool {
			peers[k.(string)] = v.(peer)
			return true
		})

		registData, err := json.Marshal(peers)
		if err != nil {
			logrus.Errorf("response data error:%v\n", err)
			return
		}

		// 发送数据，将之前保存的节点map发送给客户端
		_, err = listen.WriteToUDP(registData, addr)
		if err != nil {
			logrus.Errorf("send data error:%v\n", err)
			return
		}

		logrus.Infof("registData:%s\n", string(registData))

	}
}

func tcpServer() {
	// 建立 tcp 服务
	// listen, err := net.Listen("tcp", "0.0.0.0:9090")
	listen, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 9090,
	})
	if err != nil {
		logrus.Errorf("listen failed, err:%v\n", err)
		return
	}

	for {
		// 等待客户端建立连接
		conn, err := listen.Accept()
		if err != nil {
			logrus.Errorf("accept failed, err:%v\n", err)
			continue
		}
		// 启动一个单独的 goroutine 去处理连接
		go process(conn)
	}
}

func process(conn net.Conn) {
	// 处理完关闭连接
	defer conn.Close()

	// 针对当前连接做发送和接受操作
	for {
		reader := bufio.NewReader(conn)
		var buf [128]byte
		n, err := reader.Read(buf[:])
		if err != nil {
			logrus.Errorf("read from conn failed, err:%v\n", err)
			break
		}

		recv := string(buf[:n])
		logrus.Infof("收到的数据：%v\n", recv)

		// 将接受到的数据返回给客户端
		_, err = conn.Write([]byte("ok"))
		if err != nil {
			logrus.Errorf("write from conn failed, err:%v\n", err)
			break
		}
	}
}
