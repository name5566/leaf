package echoimpl

import (
	"github.com/name5566/leaf/demo/server/echo"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/network"
	"github.com/name5566/leaf/util"
)

// module
type Module struct {
}

func (m *Module) OnInit() {
	echo.R = util.NewCallRouter(100000)
	echo.R.Def("echo", func(args []interface{}) {
		conn := args[0].(*network.TCPConn)
		parser := args[1].(*network.MsgParser)
		data := args[2].([]byte)

		parser.Write(conn, data)
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
