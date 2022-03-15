package main

import (
	"bufio"
	"encoding/json"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
)

var peersMap = sync.Map{}
var port = 9998

type peerType = map[string]string

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
		Port: port,
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

		registerData, err := storePeers(id, addr)
		if err != nil {
			return
		}

		// 发送数据，将之前保存的节点map发送给客户端
		_, err = listen.WriteToUDP(registerData, addr)
		if err != nil {
			logrus.Errorf("send data error:%v\n", err)
			return
		}

		logrus.Infof("udp registerData:%s\n", string(registerData))

	}
}

func tcpServer() {
	// 建立 tcp 服务
	// listen, err := net.Listen("tcp", "0.0.0.0:9090")
	logrus.Info("启动 TCP server")
	listen, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: port,
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
		var buf [512]byte
		n, err := reader.Read(buf[:])
		if err != nil {
			logrus.Errorf("read from conn failed, err:%v\n", err)
			// return
			break
		}

		id := string(buf[:n])
		logrus.Infof("收到的数据：%v\n", id)

		registerData, err := storePeers(id, conn.RemoteAddr())
		if err != nil {
			logrus.Errorf("storePeers failed, err:%v\n", err)
			// return
			break
		}

		// 将接受到的数据返回给客户端
		_, err = conn.Write(registerData)
		if err != nil {
			logrus.Errorf("write from conn failed, err:%v\n", err)
			// return
			break
		}
		logrus.Infof("tcp registerData:%s\n", string(registerData))
		break //**目前的场景只做一次通信即关闭
	}
}

func storePeers(id string, addr net.Addr) (registerData []byte, err error) {
	//保存到服务器节点列表
	peer, ok := peersMap.Load(id)
	if !ok {
		peer = peerType{}
	}

	peer.(peerType)[addr.Network()] = addr.String()

	peersMap.Store(id, peer)

	//将节点信息读取到一个map中
	peers := map[string]peerType{}
	peersMap.Range(func(k interface{}, v interface{}) bool {
		peers[k.(string)] = v.(peerType)
		return true
	})

	registerData, err = json.Marshal(peers)
	if err != nil {
		logrus.Errorf("response data error:%v\n", err)
		return
	}

	return
}
