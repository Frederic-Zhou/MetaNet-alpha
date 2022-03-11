package main

import (
	"encoding/json"
	"fmt"
	"net"
)

var dialer, listener *net.UDPConn

type peer struct {
	Addr    string `json:"addr"`
	Network string `json:"network"`
}

type peerInfo struct {
	Addr    string          `json:"addr"`
	Network string          `json:"network"`
	Peers   map[string]peer `json:"peers"`
}

func dial2server() (pi peerInfo, err error) {
	dialer, err = net.DialUDP("udp",
		&net.UDPAddr{
			IP:   net.IPv4(0, 0, 0, 0),
			Port: 1234,
		},
		&net.UDPAddr{
			IP:   net.IPv4(0, 0, 0, 0),
			Port: 9090,
		})
	if err != nil {
		fmt.Printf("listen udp server error:%v\n", err)
	}
	defer dialer.Close()

	fmt.Println("will send")
	// 发送数据
	sendData := []byte("Hello server")
	_, err = dialer.Write(sendData) // 发送数据
	if err != nil {
		fmt.Println("发送数据失败，err:", err)
		return
	}

	// 接收数据
	data := make([]byte, 4096)
	n := 0
	n, _, err = dialer.ReadFromUDP(data) // 接收数据
	if err != nil {
		fmt.Println("接收数据失败，err:", err)
		return
	}

	// fmt.Println(string(data[:n]))

	err = json.Unmarshal(data[:n], &pi)
	if err != nil {
		fmt.Println("数据转换失败，err:", err)
		return
	}

	return
}

func listen4peer() {
	// 建立 udp 服务器
	listener, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 1234,
	})
	if err != nil {
		fmt.Printf("listen failed error:%v\n", err)
		return
	}
	defer listener.Close() // 使用完关闭服务

	for {
		// 接收数据
		var data [4096]byte
		n, addr, err := listener.ReadFromUDP(data[:])
		if err != nil {
			fmt.Printf("read data error:%v\n", err)
			return
		}

		fmt.Printf("addr:%v\t count:%v\t data:%v\n", addr, n, string(data[:n]))

		// 发送数据
		_, err = listener.WriteToUDP([]byte("ok"), addr)
		if err != nil {
			fmt.Printf("send data error:%v\n", err)
			return
		}
	}
}

func main() {
	pi, err := dial2server()

	fmt.Println(pi, err)
}
