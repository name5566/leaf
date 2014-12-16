package echoimpl

import (
	"github.com/name5566/leaf/demo/server/echo"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/network"
)

type Module struct{}

func (m *Module) OnInit() {
	// echo function
	echo.R.Def("echo", func(args []interface{}) {
		conn := args[0].(*network.TCPConn)
		data := args[1].([]byte)
		conn.WriteMsg(data)
	})

	log.Release("Init the Echo module")
}

func (m *Module) OnDestroy() {
	log.Release("Destroy the Echo module")
}

func (m *Module) Run(closeSig chan bool) {
	for {
		select {
		case <-closeSig:
			return
		case ci := <-echo.R.Chan():
			echo.R.Route(ci)
		}
	}
}
