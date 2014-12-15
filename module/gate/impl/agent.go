package gateimpl

import (
	"encoding/binary"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/network"
	"io"
)

type Agent struct {
	conn *network.TCPConn
	conf *AgentConf
}

type AgentConf struct {
	lenMsgLen    int
	lenMsgID     int
	maxMsgLen    uint32
	littleEndian bool
}

func readUint32(buf []byte, littleEndian bool) uint32 {
	switch len(buf) {
	case 1:
		return uint32(buf[0])
	case 2:
		if littleEndian {
			return uint32(binary.LittleEndian.Uint16(buf))
		} else {
			return uint32(binary.BigEndian.Uint16(buf))
		}
	case 4:
		if littleEndian {
			return binary.LittleEndian.Uint32(buf)
		} else {
			return binary.BigEndian.Uint32(buf)
		}
	}

	return 0
}

func (a *Agent) Run() {
	bufMsgLen := make([]byte, a.conf.lenMsgLen)
	bufMsgID := make([]byte, a.conf.lenMsgID)

	for {
		// len
		_, err := io.ReadFull(a.conn, bufMsgLen)
		if err != nil {
			log.Debug("read msg len error: %v", err)
			break
		}
		msgLen := readUint32(bufMsgLen, a.conf.littleEndian)
		if msgLen > a.conf.maxMsgLen {
			log.Debug("read msg: message too long (%v)", msgLen)
			break
		}

		// id
		_, err = io.ReadFull(a.conn, bufMsgID)
		if err != nil {
			log.Debug("read msg id error: %v", err)
			break
		}
		msgID := readUint32(bufMsgID, a.conf.littleEndian)

		// data
		msgData := make([]byte, msgLen)
		_, err = io.ReadFull(a.conn, msgData)
		if err != nil {
			log.Debug("read msg data error: %v", err)
			break
		}

		// process ...
		var _ = msgID
	}
}

func (a *Agent) OnClose() {

}
