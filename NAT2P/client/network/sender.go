package network

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/Frederic-Zhou/MetaNet-alpha/NAT2P/client/utils"
	reuse "github.com/libp2p/go-reuseport"
	"github.com/sirupsen/logrus"
)

var MAXSENDSIZE = 16
var SENDID uint32 = 0
var sendIDMutex sync.Mutex

var sendBlocks sync.Map

func getSendID() uint32 {
	sendIDMutex.Lock()
	defer sendIDMutex.Unlock()
	SENDID++
	return SENDID
}

func udpSender(reader io.Reader, laddr, raddr string) (err error) {

	logrus.Infof("向peer发送数据  %s -> %s \n", laddr, raddr)
	var dialer net.Conn
	dialer, err = reuse.Dial("udp", laddr, raddr)
	if err != nil {
		logrus.Errorf("2listen udp server error:%v\n", err)
		return
	}
	defer dialer.Close()

	var id, seq uint32 = getSendID(), 0

	for {

		var block = make([]byte, MAXSENDSIZE)
		var n = 0
		if n, err = reader.Read(block); n == 0 {
			logrus.Errorf("%v||%v", n, err)
			//读取完所有的内容，此时发送一个核对清单给接收端
			//核对信息是有效块的最大序号
			binary.LittleEndian.PutUint32(block, seq)
			seq = 0
		} else {
			seq++
		}

		check := utils.CRC32(block)

		var idbytes, seqbytes, checkbytes = make([]byte, 4), make([]byte, 4), make([]byte, 4)
		binary.LittleEndian.PutUint32(idbytes, id)
		binary.LittleEndian.PutUint32(seqbytes, seq)
		binary.LittleEndian.PutUint32(checkbytes, check)

		//保存到发送列表

		sendBlocks.Store(fmt.Sprintf("%s %d %d", raddr, id, seq), block)

		block = append(idbytes, append(seqbytes, append(checkbytes, block...)...)...)

		// 发送数据
		_, err = dialer.Write(block) // 发送数据
		if err != nil {
			logrus.Errorf("To peer 发送数据失败: %v\n", err)
			return
		}

		logrus.Infoln(">", id, seq, check, block)

		if n == 0 {
			logrus.Warnln("消息发送完成")
			return
		}
	}

}

func Sender(reader io.Reader, laddr, raddr string) (err error) {
	return udpSender(reader, laddr, raddr)
}
