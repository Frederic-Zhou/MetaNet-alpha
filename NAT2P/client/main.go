package main

import (
	"encoding/json"
	"fmt"
	"net"
)

var dialer, listener *net.UDPConn

var svcUdpAddr = net.UDPAddr{
	IP:   net.IPv4(0, 0, 0, 0),
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
		fmt.Printf("1.listen udp server error:%v\n", err)
	}
	defer dialer.Close()

	fmt.Println("will send to sever")
	// 发送数据
	sendData := []byte("Hello server")
	_, err = dialer.Write(sendData) // 发送数据
	if err != nil {
		fmt.Println("1发送数据失败，err:", err)
		return
	}

	// 接收数据
	data := make([]byte, 4096)
	n := 0
	n, _, err = dialer.ReadFromUDP(data) // 接收数据
	if err != nil {
		fmt.Println("1接收数据失败，err:", err)
		return
	}

	err = json.Unmarshal(data[:n], &rr.Peers)
	if err != nil {
		fmt.Println("1数据转换失败，err:", err)
		return
	}

	rr.LocalAddr, err = net.ResolveUDPAddr(dialer.LocalAddr().Network(), dialer.LocalAddr().String())

	return
}

func udpListen4Peer(laddr *net.UDPAddr) (err error) {

	fmt.Println("listen", laddr.String())
	// 建立 udp 服务器
	listener, err = net.ListenUDP("udp", laddr)
	if err != nil {
		fmt.Printf("listen failed error:%v\n", err)
		return
	}
	defer listener.Close() // 使用完关闭服务

	for {
		fmt.Println("listen 等待接收数据")
		// 接收数据
		var data [4096]byte
		var addr *net.UDPAddr
		var n int
		n, addr, err = listener.ReadFromUDP(data[:])
		if err != nil {
			fmt.Printf("read data error:%v\n", err)
			return
		}
		fmt.Println("listen 接收到一个数据")
		fmt.Printf("addr:%v\t count:%v\t data:%v\n", addr, n, string(data[:n]))

		// 发送数据
		_, err = listener.WriteToUDP([]byte("ok"), addr)
		if err != nil {
			fmt.Printf("send data error:%v\n", err)
			return
		}
	}
}

func udpSendmsg2Peer(msg string, laddr, raddr *net.UDPAddr) (err error) {

	fmt.Printf("向peer发送数据 \"%s\" %s -> %s \n", msg, laddr.String(), raddr.String())
	dialer, err = net.DialUDP("udp", laddr, raddr)
	if err != nil {
		fmt.Printf("2listen udp server error:%v\n", err)
	}
	defer dialer.Close()

	// 发送数据
	sendData := []byte(msg)
	_, err = dialer.Write(sendData) // 发送数据
	if err != nil {
		fmt.Println("2发送数据失败，err:", err)
		return
	}

	fmt.Print("等待回收数据...:")
	// 接收数据
	data := make([]byte, 4096)
	n := 0
	n, _, err = dialer.ReadFromUDP(data) // 接收数据
	if err != nil {
		fmt.Println("2接收数据失败，err:", err)
		return
	}

	fmt.Println(string(data[:n]))

	return
}

func main() {

	//1.与服务器通信，并获得
	rr, err := udpDial2Server(&svcUdpAddr)

	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(rr, err)

	//2.向所有peer发送UDP请求，打通隧道
	for name, addr := range rr.Peers {
		raddr, err := net.ResolveUDPAddr(addr.Network, addr.Addr)
		if err != nil {
			fmt.Println(err)
			continue
		}
		err = udpSendmsg2Peer(name, rr.LocalAddr.(*net.UDPAddr), raddr)
		if err != nil {
			fmt.Println(err)
		}
	}

	//3.监听刚才与服务器通信的本地端口
	udpListen4Peer(rr.LocalAddr.(*net.UDPAddr))

}
