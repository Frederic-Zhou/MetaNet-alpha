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
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var PKGSIZE = BLOCKSIZE + 12 //4位ID+4位序号+4位校验=12，所以接收端收到的每个块会多12位
var receiveCache *leveldb.DB

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

		//校验错误
		if check != utils.CRC32(block) {
			logrus.Infof("校验错误 数据校验结果 %v !=  传入的校验值  %v", utils.CRC32(block), check)
			//如果收到校验错误的数据块，跳过下面的保存缓冲的操作，丢弃！
			continue
		}

		if seq == 0 {
			lastseq, datatype := getLastseqAndDataType(binary.LittleEndian.Uint32(block))
			logrus.Infof("收到结束消息 %v,%v,%v", seq, lastseq, datatype)
			go fetchReceiveCache(raddr.String(), id, lastseq, datatype)
		} else {
			go func() {
				err = receiveCache.Put([]byte(fmt.Sprintf("%s-%d-%d", raddr.String(), id, seq)), block, nil)
				if err != nil {
					logrus.Errorf("write to cache:%v", err)
				}
			}()
		}

		logrus.Infof("来源:%v > %d-%d-%d-%v", raddr, id, seq, check, block)

	}
}

func fetchReceiveCache(raddr string, sendid, lastseq, datatype uint32) {
	time.Sleep(1 * time.Second)
	logrus.Infoln("fetch:", fmt.Sprintf("%s-%d", raddr, sendid))

	iterRange := &util.Range{
		Start: []byte(fmt.Sprintf("%s-%d-%d", raddr, sendid, 1)),
		Limit: []byte(fmt.Sprintf("%s-%d-%d", raddr, sendid, lastseq+1)),
	}
	iter := receiveCache.NewIterator(iterRange, nil)

	//检查完整性
	if err := checkCache(receiveCache.NewIterator(iterRange, nil), lastseq); err != nil && datatype != DataType_Success {
		logrus.Errorln(err)
		return
	}

	//查询
	switch datatype {
	case DataType_Success: //成功消息
		logrus.Infoln("收到成功回报")

		//如果是成功回报，lastseq实际是回报的sendid
		//清理发送缓存
		successSendID := lastseq
		delIter := sendCache.NewIterator(
			util.BytesPrefix([]byte(fmt.Sprintf("%s-%d-", raddr, successSendID))),
			nil)

		for delIter.Next() {
			sendCache.Delete(delIter.Key(), nil)
		}

		delIter.Release()
		if err := delIter.Error(); err != nil {
			logrus.Errorf("del cache err: %v", err)
		}

		logrus.Infoln("清理完成")

	case DataType_Text: //文本消息
		logrus.Infoln("组装消息")
		fullData := []byte{}
		for iter.Next() {
			value := iter.Value()
			fullData = append(fullData, value...)
			//清理缓存数据
			receiveCache.Delete(iter.Key(), nil)
		}

		logrus.Infof("Text:%v\n", string(fullData))

		//向发送方发送接收成功消息
		var sendidbyte = make([]byte, 4)
		binary.LittleEndian.PutUint32(sendidbyte, sendid)
		Sender(bytes.NewReader(sendidbyte), DataType_Success, "0.0.0.0:9997", raddr)
		logrus.Infof("发送成功回报:%v > %v\n", sendidbyte, raddr)

	case DataType_File:

	}

	iter.Release()
	if err := iter.Error(); err != nil {
		logrus.Errorf("receive cache err: %v", err)
	}
}

func checkCache(iter iterator.Iterator, lastseq uint32) (err error) {
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

func getLastseqAndDataType(code uint32) (lastseq, datatype uint32) {
	return code / 10, code % 10
}

func Listener(laddr string) (err error) {
	return udpListen4Peers(laddr)
}
