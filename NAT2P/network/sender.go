package network

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"sync"

	"github.com/Frederic-Zhou/MetaNet-alpha/NAT2P/utils"
	reuse "github.com/libp2p/go-reuseport"
	"github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

var BLOCKSIZE = 32
var SENDID uint32 = 0
var sendIDMutex sync.Mutex

var sendCache *leveldb.DB

type DataType = uint32

const (
	DataType_Reply DataType = 0
	DataType_Text  DataType = 1
	DataType_File  DataType = 2
	DataType_Flow  DataType = 3
)

func init() {
	var err error
	var cachePath = "./sendCache"
	err = os.RemoveAll(cachePath)
	if err != nil {
		logrus.Errorln(err)
		os.Exit(4)
	}
	sendCache, err = leveldb.OpenFile(cachePath, nil)
	if err != nil {
		os.Exit(4)
	}
}

func getSendID() uint32 {
	sendIDMutex.Lock()
	defer sendIDMutex.Unlock()
	SENDID++
	return SENDID
}

func udpSender(reader io.Reader, datatype DataType, laddr, raddr string) (err error) {
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

		var block = make([]byte, BLOCKSIZE)
		var n = 0
		//如何读取完成或者**发送的是成功回报** ，那么发送seq=0，代表结束数据，不会有后续内容了
		if n, err = reader.Read(block); n == 0 || datatype == DataType_Reply {
			logrus.Errorf("读取: %v, 错误: %v, 数据类型: %v", n, err, datatype)
			//读取完所有的内容，此时发送一个核对清单给接收端
			//核对信息是有效块的最大序号 *10 + 发送类型
			//如果是**成功回报** 正常发送
			switch datatype {
			case DataType_Reply:
				//成功回报
				copy(block, genEndInfo(seq, datatype, block[:]))
			case DataType_Text:
				//消息发送结束
				copy(block, genEndInfo(seq, datatype, []byte{}))

			case DataType_File:

			}

			seq = 0
		} else {
			seq++
		}

		check := utils.CRC32(block)

		var idbytes, seqbytes, checkbytes = make([]byte, 4), make([]byte, 4), make([]byte, 4)
		binary.LittleEndian.PutUint32(idbytes, id)
		binary.LittleEndian.PutUint32(seqbytes, seq)
		binary.LittleEndian.PutUint32(checkbytes, check)

		pkg := append(idbytes, append(seqbytes, append(checkbytes, block...)...)...)

		// 发送数据
		_, err = dialer.Write(pkg) // 发送数据
		if err != nil {
			logrus.Errorf("To peer 发送数据失败: %v\n", err)
			return
		}

		if seq != 0 {
			//如果发送的不是成功回报，保存到发送列表
			go func() {
				err := sendCache.Put([]byte(fmt.Sprintf("%s-%d-%d", raddr, id, seq)), block, nil)
				if err != nil {
					logrus.Errorf("write to cache:%v", err)
				}
			}()
		} else {
			logrus.Warnln("消息发送完成")
			return
		}

		logrus.Infoln(">", id, seq, check, pkg)
	}

}

func Sender(reader io.Reader, datatype DataType, laddr, raddr string) (err error) {
	return udpSender(reader, datatype, laddr, raddr)
}
