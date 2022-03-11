package main

import (
	"fmt"
	"net"
)

func main() {
	// 建立服务
	listen, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(1, 14, 102, 100),
		Port: 9090,
	})
	if err != nil {
		fmt.Printf("listen udp server error:%v\n", err)
	}
	defer listen.Close()

	fmt.Println("will send")
	// 发送数据
	sendData := []byte("Hello server")
	_, err = listen.Write(sendData) // 发送数据
	if err != nil {
		fmt.Println("发送数据失败，err:", err)
		return
	}

	fmt.Println("sended")

	// 接收数据
	data := make([]byte, 4096)
	n, remoteAddr, err := listen.ReadFromUDP(data) // 接收数据
	if err != nil {
		fmt.Println("接收数据失败，err:", err)
		return
	}

	fmt.Println("readed")
	fmt.Printf("recv:%v addr:%v count:%v\n", string(data[:n]), remoteAddr, n)
	fmt.Println(listen.LocalAddr().String())
}
