package extra

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/util"
	"math"
	"reflect"
)

// -------------------------
// | id | protobuf message |
// -------------------------
type ProtobufRouter struct {
	littleEndian bool
	msgInfo      []*ProtobufMsgInfo
	msgID        map[reflect.Type]uint16
}

type ProtobufMsgInfo struct {
	msgType    reflect.Type
	msgRouter  *util.CallRouter
	msgHandler ProtobufMsgHandler
}

type ProtobufMsgHandler func([]interface{})

func NewProtobufRouter() *ProtobufRouter {
	r := new(ProtobufRouter)
	r.littleEndian = false
	r.msgID = make(map[reflect.Type]uint16)
	return r
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
func (r *ProtobufRouter) SetByteOrder(littleEndian bool) {
	r.littleEndian = littleEndian
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
func (r *ProtobufRouter) RegisterRouter(msg proto.Message, msgRouter *util.CallRouter) {
	r.protobufMsgInfo(msg).msgRouter = msgRouter
}

// It's dangerous to call the method on routing or marshaling (unmarshaling)
func (r *ProtobufRouter) RegisterHandler(msg proto.Message, msgHandler ProtobufMsgHandler) {
	r.protobufMsgInfo(msg).msgHandler = msgHandler
}

func (r *ProtobufRouter) protobufMsgInfo(msg proto.Message) *ProtobufMsgInfo {
	msgType := reflect.TypeOf(msg)
	if msgType == nil || msgType.Kind() != reflect.Ptr {
		log.Fatal("protobuf message pointer required")
	}

	if id, ok := r.msgID[msgType]; ok {
		return r.msgInfo[id]
	}

	if len(r.msgInfo) >= math.MaxUint16 {
		log.Fatal("too many protobuf messages (max = %v)", math.MaxUint16)
	}

	i := new(ProtobufMsgInfo)
	i.msgType = msgType
	r.msgInfo = append(r.msgInfo, i)
	r.msgID[msgType] = uint16(len(r.msgInfo) - 1)
	return i
}

// goroutine safe
func (r *ProtobufRouter) Route(msg proto.Message, userData interface{}) error {
	msgType := reflect.TypeOf(msg)

	id, ok := r.msgID[msgType]
	if !ok {
		return errors.New(fmt.Sprintf("message %s not registered", msgType))
	}

	i := r.msgInfo[id]
	if i.msgHandler != nil {
		i.msgHandler([]interface{}{msgType, msg, userData})
	}
	if i.msgRouter != nil {
		i.msgRouter.AsynCall0(msgType, msg, userData)
	}
	return nil
}

// goroutine safe
func (r *ProtobufRouter) Unmarshal(data []byte) (proto.Message, error) {
	if len(data) < 2 {
		return nil, errors.New("protobuf data too short")
	}

	// id
	var id uint16
	if r.littleEndian {
		id = binary.LittleEndian.Uint16(data)
	} else {
		id = binary.BigEndian.Uint16(data)
	}

	// msg
	if id >= uint16(len(r.msgInfo)) {
		return nil, errors.New(fmt.Sprintf("message id %v not registered", id))
	}
	msg := reflect.New(r.msgInfo[id].msgType.Elem()).Interface().(proto.Message)
	return msg, proto.UnmarshalMerge(data[2:], msg)
}

// goroutine safe
func (r *ProtobufRouter) Marshal(msg proto.Message) (id []byte, data []byte, err error) {
	msgType := reflect.TypeOf(msg)

	// id
	_id, ok := r.msgID[msgType]
	if !ok {
		err = errors.New(fmt.Sprintf("message %s not registered", msgType))
		return
	}

	id = make([]byte, 2)
	if r.littleEndian {
		binary.LittleEndian.PutUint16(id, _id)
	} else {
		binary.BigEndian.PutUint16(id, _id)
	}

	// data
	data, err = proto.Marshal(msg)
	return
}
