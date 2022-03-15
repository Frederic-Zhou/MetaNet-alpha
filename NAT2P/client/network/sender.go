package network

import (
	"io"
	"net"

	reuse "github.com/libp2p/go-reuseport"
	"github.com/sirupsen/logrus"
)

func udpSender(reader io.Reader, laddr, raddr string) (err error) {

	logrus.Infof("向peer发送数据  %s -> %s \n", laddr, raddr)
	var dialer net.Conn
	dialer, err = reuse.Dial("udp", laddr, raddr)
	if err != nil {
		logrus.Errorf("2listen udp server error:%v\n", err)
		return
	}
	defer dialer.Close()

	for {
		var data = make([]byte, MAXSIZE)
		var n = 0
		if n, err = reader.Read(data); err != nil {
			logrus.Errorf("%v||%v", n, err)
			return
		}

		// 发送数据
		_, err = dialer.Write(data) // 发送数据
		if err != nil {
			logrus.Errorf("To peer 发送数据失败: %v\n", err)
			return
		}

		logrus.Infoln(">", string(data[:]))

	}

}

func Sender(reader io.Reader, laddr, raddr string) (err error) {
	return udpSender(reader, laddr, raddr)
}
