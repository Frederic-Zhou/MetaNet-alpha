package network

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"sync"

	"github.com/Frederic-Zhou/MetaNet-alpha/NAT2P/client/utils"
	reuse "github.com/libp2p/go-reuseport"
	"github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

var MAXSENDSIZE = 16
var SENDID uint32 = 0
var sendIDMutex sync.Mutex

var sendCache *leveldb.DB

type DataType = uint32

const (
	DataType_Success DataType = 0
	DataType_Text    DataType = 1
	DataType_File    DataType = 2
	DataType_Image   DataType = 3
	DataType_Flow    DataType = 4
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

		var block = make([]byte, MAXSENDSIZE)
		var n = 0
		//如何读取完成或者**发送的是成功回报** ，那么发送seq=0，代表结束数据，不会有后续内容了
		if n, err = reader.Read(block); n == 0 || datatype == DataType_Success {
			logrus.Errorf("%v||%v||%v", n, err, datatype)
			//读取完所有的内容，此时发送一个核对清单给接收端
			//核对信息是有效块的最大序号 *10 + 发送类型
			//如果是**成功回报** 正常发送
			if datatype == DataType_Success {
				//成功回报，将sql替换为要回报sendid
				binary.LittleEndian.PutUint32(block, genLastseqAndDataType(binary.LittleEndian.Uint32(block), datatype))
			} else {
				binary.LittleEndian.PutUint32(block, genLastseqAndDataType(seq, datatype))
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

		if datatype != DataType_Success {
			//如果发送的不是成功的回报，保存到发送列表
			go func() {
				err := sendCache.Put([]byte(fmt.Sprintf("%s-%d-%d", raddr, id, seq)), block, nil)
				if err != nil {
					logrus.Errorf("write to cache:%v", err)
				}
			}()
		}

		block = append(idbytes, append(seqbytes, append(checkbytes, block...)...)...)

		// 发送数据
		_, err = dialer.Write(block) // 发送数据
		if err != nil {
			logrus.Errorf("To peer 发送数据失败: %v\n", err)
			return
		}

		logrus.Infoln(">", id, seq, check, block)

		if seq == 0 {
			logrus.Warnln("消息发送完成")
			return
		}
	}

}

func genLastseqAndDataType(lastseq, datatype uint32) uint32 {
	return lastseq*10 + datatype
}

func Sender(reader io.Reader, datatype DataType, laddr, raddr string) (err error) {
	return udpSender(reader, datatype, laddr, raddr)
}
