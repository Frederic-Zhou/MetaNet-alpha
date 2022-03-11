package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sync"
)

var peersMap = sync.Map{}

type peer struct {
	Addr    string `json:"addr"`
	Network string `json:"network"`
}

func main() {
	udpServer()
	// tcpServer()
}

func udpServer() {
	// 建立 udp 服务器
	listen, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 9090,
	})
	if err != nil {
		fmt.Printf("listen failed error:%v\n", err)
		return
	}
	defer listen.Close() // 使用完关闭服务

	for {
		// 接收数据
		var data [1024]byte
		var id string
		n, addr, err := listen.ReadFromUDP(data[:])
		if err != nil {
			fmt.Printf("read data error:%v\n", err)
			return
		}

		id = string(data[:n])
		fmt.Printf("addr:%v\t count:%v\t data:%v\n", addr, n, string(data[:n]))

		peersMap.Store(id, peer{Addr: addr.String(), Network: addr.Network()})

		peers := map[string]peer{}
		peersMap.Range(func(k interface{}, v interface{}) bool {
			peers[k.(string)] = v.(peer)
			return true
		})
		responsedata, err := json.Marshal(map[string]interface{}{
			"addr":    addr.String(),
			"network": addr.Network(),
			"peers":   peers,
		})

		fmt.Println(string(responsedata))

		if err != nil {
			fmt.Printf("response data error:%v\n", err)
			return
		}

		// 发送数据
		_, err = listen.WriteToUDP(responsedata, addr)
		if err != nil {
			fmt.Printf("send data error:%v\n", err)
			return
		}
	}
}

func tcpServer() {
	// 建立 tcp 服务
	listen, err := net.Listen("tcp", "0.0.0.0:9090")
	if err != nil {
		fmt.Printf("listen failed, err:%v\n", err)
		return
	}

	for {
		// 等待客户端建立连接
		conn, err := listen.Accept()
		if err != nil {
			fmt.Printf("accept failed, err:%v\n", err)
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
			fmt.Printf("read from conn failed, err:%v\n", err)
			break
		}

		recv := string(buf[:n])
		fmt.Printf("收到的数据：%v\n", recv)

		// 将接受到的数据返回给客户端
		_, err = conn.Write([]byte("ok"))
		if err != nil {
			fmt.Printf("write from conn failed, err:%v\n", err)
			break
		}
	}
}
