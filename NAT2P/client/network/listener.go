package network

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/Frederic-Zhou/MetaNet-alpha/NAT2P/client/utils"
	reuse "github.com/libp2p/go-reuseport"
	"github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var PKGSIZE = BLOCKSIZE + 12 //4位ID+4位序号+4位校验=12，所以接收端收到的每个块会多12位
var receiveCache *leveldb.DB
var LADDR = ""

func init() {
	var err error
	var cachePath = "./receiveCache"
	err = os.RemoveAll(cachePath)
	if err != nil {
		logrus.Errorln(err)
		os.Exit(4)
	}
	receiveCache, err = leveldb.OpenFile(cachePath, nil)
	if err != nil {
		logrus.Errorln(err)
		os.Exit(4)
	}
}

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
		var pkg = make([]byte, PKGSIZE)
		var raddr net.Addr

		_, raddr, err = listenerUDP.ReadFrom(pkg)
		if err != nil {
			logrus.Errorf("读取数据错误:%v\n", err)
			return
		}

		var idbytes, seqbytes, checkbytes, block = make([]byte, 4), make([]byte, 4), make([]byte, 4), []byte{}

		idbytes = pkg[:4]
		seqbytes = pkg[4:8]
		checkbytes = pkg[8:12]
		block = pkg[12:]

		id := binary.LittleEndian.Uint32(idbytes)
		seq := binary.LittleEndian.Uint32(seqbytes)
		check := binary.LittleEndian.Uint32(checkbytes)

		logrus.Infof("来源:%v > %d-%d-%d-%v", raddr, id, seq, check, block)
		//校验错误
		if check != utils.CRC32(block) {
			logrus.Infof("校验错误 数据校验结果 %v !=  传入的校验值  %v , %v ", utils.CRC32(block), check, block)
			//如果收到校验错误的数据块，跳过下面的保存缓冲的操作，丢弃！
			continue
		}

		if seq == 0 { //seq=0, 要么是消息发送结束或成功回报，不进入缓冲区，并且立即回调处理缓存的数据
			lastseq, datatype, endInfo := parseEndInfo(block)
			logrus.Infof("收到结束消息 %v,%v,%v", lastseq, datatype, endInfo)
			go fetchReceiveCache(raddr.String(), id, lastseq, datatype, endInfo)
		} else {
			go func() {
				err = receiveCache.Put([]byte(fmt.Sprintf("%s-%d-%d", raddr.String(), id, seq)), block, nil)
				if err != nil {
					logrus.Errorf("write to cache:%v", err)
				}
			}()
		}

	}
}

func fetchReceiveCache(raddr string, sendid, lastseq, datatype uint32, endInfo []byte) {

	if datatype == DataType_Success {
		handleSendSuccess(binary.LittleEndian.Uint32(endInfo), raddr)
		return
	}

	time.Sleep(1 * time.Second) // 临时举措
	iterRange := &util.Range{
		Start: []byte(fmt.Sprintf("%s-%d-%d", raddr, sendid, 1)),
		Limit: []byte(fmt.Sprintf("%s-%d-%d", raddr, sendid, lastseq+1)),
	}
	logrus.Infoln("fetch:", fmt.Sprintf("%s-%d", raddr, sendid))

	defer sendSuccess(sendid, raddr) //发送成功回复给发送方

	//检查数据是否完整
	if err := checkCache(iterRange, lastseq); err != nil {
		logrus.Errorln(err)
		return
	}

	//查询
	switch datatype {
	case DataType_Text:
		_, _ = handleText(iterRange)
	case DataType_File:

	}

}

func sendSuccess(sendid uint32, raddr string) {
	var sendidbyte = make([]byte, 4)
	binary.LittleEndian.PutUint32(sendidbyte, sendid)
	go Sender(bytes.NewReader(sendidbyte), DataType_Success, LADDR, raddr)
	logrus.Infof("发送成功回报:%v > %v\n", sendidbyte, raddr)
}

func handleText(iterRange *util.Range) (text string, err error) {

	iter := receiveCache.NewIterator(iterRange, nil)

	//从缓存中读取数据到字节数组，并且删除数据
	fullData := []byte{}
	for iter.Next() {
		value := iter.Value()
		fullData = append(fullData, value...)
		receiveCache.Delete(iter.Key(), nil)
	}

	text = string(fullData)
	logrus.Infof("Text:%v\n", text)

	iter.Release()
	if err = iter.Error(); err != nil {
		logrus.Errorf("receive cache err: %v", err)
		return
	}

	return
}

func handleSendSuccess(pkgInfo uint32, raddr string) {
	logrus.Infoln("是成功回报")

	successSendID := pkgInfo
	delIter := sendCache.NewIterator(
		util.BytesPrefix([]byte(fmt.Sprintf("%s-%d-", raddr, successSendID))), nil)

	for delIter.Next() {
		sendCache.Delete(delIter.Key(), nil)
	}

	delIter.Release()
	if err := delIter.Error(); err != nil {
		logrus.Errorf("del cache err: %v", err)
	}

	logrus.Infoln("清理完成")
}

//临时举措
func checkCache(iterRange *util.Range, lastseq uint32) (err error) {

	iter := receiveCache.NewIterator(iterRange, nil)

	var countSeq uint32 = 0
	for iter.Next() {
		countSeq++
	}

	if countSeq != lastseq {
		err = fmt.Errorf("消息不完整 %v/%v", countSeq, lastseq)
	}

	iter.Release()
	if err := iter.Error(); err != nil {
		logrus.Errorf("check cache err: %v", err)
	}

	return
}

func genEndInfo(lastseq, datatype uint32, endinfo []byte) []byte {

	var lastseqByte, datatypeByte = make([]byte, 4), make([]byte, 4)
	binary.LittleEndian.PutUint32(lastseqByte, lastseq)
	binary.LittleEndian.PutUint32(datatypeByte, datatype)

	var block = make([]byte, BLOCKSIZE)
	blockReader := bytes.NewReader(append(lastseqByte, append(datatypeByte, endinfo...)...))
	_, _ = blockReader.Read(block)
	return block
}

func parseEndInfo(block []byte) (lastseq, datatype uint32, endinfo []byte) {

	lastseq = binary.LittleEndian.Uint32(block[:4])
	datatype = binary.LittleEndian.Uint32(block[4:8])
	endinfo = block[8:]
	return
}

func Listener(laddr string) (err error) {
	LADDR = laddr
	return udpListen4Peers(laddr)
}
