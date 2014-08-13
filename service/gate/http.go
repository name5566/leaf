package gate

import (
	"errors"
)

type HttpGateConfig struct {
}

type HttpGate struct {
}

func NewHttpGate(config *HttpGateConfig) (*HttpGate, error) {
	return nil, errors.New("not implemented")
}

func (httpGate *HttpGate) Start() {
}

func (httpGate *HttpGate) Close() {
}
