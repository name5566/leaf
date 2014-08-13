package gate

import (
	"errors"
)

type HttpGateCfg struct {
}

type HttpGate struct {
}

func NewHttpGate(cfg *HttpGateCfg) (*HttpGate, error) {
	return nil, errors.New("not implemented")
}

func (httpGate *HttpGate) Start() {
}

func (httpGate *HttpGate) Close() {
}
