package network

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"

	"github.com/Frederic-Zhou/MetaNet-alpha/NAT2P/client/utils"
	reuse "github.com/libp2p/go-reuseport"
	"github.com/sirupsen/logrus"
)

var MAXRECEIVESIZE = MAXSENDSIZE + 12 //4位ID+4位序号+4位校验=12，所以接收端收到的每个块会多12位
var receiveBlocks sync.Map

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

		// 接收数据
		var block = make([]byte, MAXRECEIVESIZE)
		var raddr net.Addr

		_, raddr, err = listenerUDP.ReadFrom(block)
		if err != nil {
			logrus.Errorf("读取数据错误:%v\n", err)
			return
		}

		var idbytes, seqbytes, checkbytes = make([]byte, 4), make([]byte, 4), make([]byte, 4)

		idbytes = block[:4]
		seqbytes = block[4:8]
		checkbytes = block[8:12]
		block = block[12:]

		id := binary.LittleEndian.Uint32(idbytes)
		seq := binary.LittleEndian.Uint32(seqbytes)
		check := binary.LittleEndian.Uint32(checkbytes)

		//校验错误
		if check != utils.CRC32(block) {
			logrus.Infof("校验错误 数据校验结果 %v !=  传入的校验值  %v", utils.CRC32(block), check)
		}

		if seq == 0 {
			maxseq := binary.LittleEndian.Uint32(block)
			logrus.Infof("收到结束消息 %v,%v", seq, maxseq)
		}

		receiveBlocks.Store(fmt.Sprintf("%s %d %d", raddr.String(), id, seq), block)

		logrus.Infof("来源:%v > %d %d %d %v", raddr, id, seq, check, block)

	}
}

func Listern(laddr string) (err error) {
	return udpListen4Peers(laddr)
}
