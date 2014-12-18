package extra

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/name5566/leaf/log"
	"math"
	"reflect"
)

// -------------------------
// | id | protobuf message |
// -------------------------
type ProtobufParser struct {
	littleEndian bool
	msgType      []reflect.Type
	msgID        map[reflect.Type]uint16
}

func NewProtobufParser() *ProtobufParser {
	p := new(ProtobufParser)
	p.msgID = make(map[reflect.Type]uint16)
	return p
}

// It's dangerous to call the method on marshaling or unmarshaling
func (p *ProtobufParser) SetByteOrder(littleEndian bool) {
	p.littleEndian = littleEndian
}

// It's dangerous to call the method on marshaling or unmarshaling
func (p *ProtobufParser) Register(msg proto.Message) {
	if len(p.msgType) >= math.MaxUint16 {
		log.Fatal("too many protobuf messages (max = %v)", math.MaxUint16)
	}

	t := reflect.TypeOf(msg)
	if t == nil || t.Kind() != reflect.Ptr {
		log.Fatal("protobuf message pointer required")
	}

	p.msgType = append(p.msgType, t)
	p.msgID[t] = uint16(len(p.msgType) - 1)
}

// goroutine safe
func (p *ProtobufParser) Unmarshal(data []byte) (id uint16, msg proto.Message, err error) {
	if len(data) < 2 {
		err = errors.New("protobuf data too short")
		return
	}

	// id
	if p.littleEndian {
		id = binary.LittleEndian.Uint16(data)
	} else {
		id = binary.BigEndian.Uint16(data)
	}

	// msg
	if id >= uint16(len(p.msgType)) {
		err = errors.New(fmt.Sprintf("message id %v not registered", id))
		return
	}
	t := p.msgType[id]
	msg = reflect.New(t.Elem()).Interface().(proto.Message)
	err = proto.UnmarshalMerge(data[2:], msg)
	return
}

// goroutine safe
func (p *ProtobufParser) Marshal(msg proto.Message) (id uint16, data []byte, err error) {
	t := reflect.TypeOf(msg)

	// id
	if _id, ok := p.msgID[t]; !ok {
		err = errors.New(fmt.Sprintf("message %s not registered", t))
		return
	} else {
		id = _id
	}

	// data
	data, err = proto.Marshal(msg)
	return
}
