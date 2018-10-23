package flatbuffer

import (
	"reflect"
	"github.com/name5566/leaf/chanrpc"
	"github.com/google/flatbuffers/go"
	"github.com/name5566/leaf/log"
	"math"
	"fmt"
	"encoding/binary"
	"github.com/pkg/errors"
)

type Processor struct{
	littleEndian 	bool
	msgInfo 		[]*MsgInfo
	msgID			map[reflect.Type]uint16
}

type MsgInfo struct{
	msgType 		reflect.Type
	msgRouter		*chanrpc.Server
	msgHandler 		MsgHandler
	msgRawHandler	MsgHandler
}

type MsgHandler func([]interface{})

type MsgRaw struct{
	msgID 		uint16
	msgRawData 	[]byte
}

func NewProcessor() *Processor{
	p := new(Processor)
	p.littleEndian = false
	p.msgID = make(map[reflect.Type]uint16)
	return p
}

func (p *Processor) SetByteOrder(littleENdian bool){
	p.littleEndian = littleENdian
}

func (p *Processor) Register(msgM flatbuffers.FlatBuffer) uint16 {
	//proto.Message
	msgType := reflect.TypeOf(msgM)
	if msgType == nil || msgType.Kind() != reflect.Ptr {
		log.Fatal("flatbuffers table pointer required")
	}

	if _, ok := p.msgID[msgType]; ok {
		log.Fatal("table %s is already registered", msgType)
	}

	if len(p.msgInfo) >= math.MaxUint16 {
		log.Fatal("too many flatbuffers tables (max = %v)", math.MaxUint16)
	}

	i := new(MsgInfo)
	i.msgType = msgType
	p.msgInfo = append(p.msgInfo, i)
	id := uint16(len(p.msgInfo) - 1)
	p.msgID[msgType] = id
	return id
}

func (p *Processor) SetRouter(msg flatbuffers.FlatBuffer, msgRouter *chanrpc.Server) {
	msgType := reflect.TypeOf(msg)
	id, ok := p.msgID[msgType]
	if !ok {
		log.Fatal("message %s not registered", msgType)
	}

	p.msgInfo[id].msgRouter = msgRouter
}

func (p *Processor) SetHandler(msg flatbuffers.Table, msgHandler MsgHandler) {
	msgType := reflect.TypeOf(msg)
	id, ok := p.msgID[msgType]
	if !ok {
		log.Fatal("message %s not registered", msgType)
	}

	p.msgInfo[id].msgHandler = msgHandler
}

func (p *Processor) SetRawHandler(id uint16, msgRawHandler MsgHandler) {
	if id >= uint16(len(p.msgInfo)) {
		log.Fatal("message id %v not registered", id)
	}

	p.msgInfo[id].msgRawHandler = msgRawHandler
}

func (p *Processor) RawRoute(args []interface{}) {
	msgID := args[0].(uint16)
	//msgRawData := args[1].([]byte)
	//p.msgInfo[msgID]
	i := p.msgInfo[msgID]
	if i == nil {
		fmt.Errorf("json.go RawRoute: message %v not registered", msgID)
	}


	// {msgID string, msg interface{}, userData interface{}}
	//i, ok := p.msgInfo[msgID]
	//msgType := reflect.TypeOf(msg)
	//if msgType == nil || msgType.Kind() != reflect.Ptr {
	//	return errors.New("json message pointer required")
	//}
	//msgID := msgType.Elem().Name()
	//i, ok := p.msgInfo[msgID]
	//if !ok {
	//	return fmt.Errorf("message %v not registered", msgID)
	//}
	//if i.msgHandler != nil {
	//	i.msgHandler([]interface{}{msg, userData})
	//}
	//
	if i.msgRouter != nil {
		i.msgRouter.Go(i.msgType, args[1], args[2])
	}
	//
}

func (p *Processor) Route(msg interface{}, userData interface{}) error {
	if msgRaw, ok := msg.(MsgRaw); ok {
		if msgRaw.msgID >= uint16(len(p.msgInfo)) {
			return fmt.Errorf("message id %v not registered", msgRaw.msgID)
		}

		i := p.msgInfo[msgRaw.msgID]
		if i.msgRawHandler != nil {
			i.msgRawHandler([]interface{}{msgRaw.msgID, msgRaw.msgRawData, userData})
		}
		return nil
	}

	// flatbuffers
	msgType := reflect.TypeOf(msg)
	id, ok := p.msgID[msgType]
	if !ok {
		return fmt.Errorf("message %s not registered", msgType)
	}

	i := p.msgInfo[id]
	if i.msgHandler != nil {
		i.msgHandler([]interface{}{msg, userData})
	}

	if i.msgRouter != nil {
		i.msgRouter.Go(msgType, msg, userData)
	}

	return nil
}

func (p *Processor) Unmarshal(data []byte) (interface{}, error) {
	if len(data) < 2 {
		return nil, errors.New("flatbuffer data too short")
	}

	// id
	var id uint16
	if p.littleEndian {
		id = binary.LittleEndian.Uint16(data)
	} else {
		id = binary.BigEndian.Uint16(data)
	}

	if id > uint16(len(p.msgInfo)) {
		return MsgRaw{id, data[2:]}, nil
	}

	// msg
	i := p.msgInfo[id]
	if i.msgRawHandler != nil {
		return MsgRaw{id, data[2:]}, nil
	} else {
		msg := reflect.New(i.msgType.Elem()).Interface()
		return msg, nil
	}
}

func (p *Processor) Marshal(msg interface{})([][]byte, error) {
	//log.Debug("flatbuffer.go:Marshal")
	msgType := reflect.TypeOf(msg)

	// id
	_id, ok := p.msgID[msgType]
	if !ok {
		err := fmt.Errorf("message %s not registered", msgType)
		return nil, err
	}

	id := make([]byte, 2)
	if p.littleEndian {
		binary.LittleEndian.PutUint16(id, _id)
	} else {
		binary.BigEndian.PutUint16(id, _id)
	}

	data := msg.(flatbuffers.FlatBuffer)
	//data, err := proto.Marshal(msg.(proto.Message))
	return [][]byte{id, data.Table().Bytes}, nil
}

func (p *Processor) Range(f func(id uint16, t reflect.Type)) {
	for id, i := range p.msgInfo {
		f(uint16(id), i.msgType)
	}
}

