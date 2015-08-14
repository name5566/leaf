package console

import (
	"bufio"
	"github.com/name5566/leaf/conf"
	"github.com/name5566/leaf/network"
	"math"
	"strconv"
	"strings"
)

var server *network.TCPServer

func Init() {
	if conf.ConsolePort == 0 {
		return
	}

	server = new(network.TCPServer)
	server.Addr = "localhost:" + strconv.Itoa(conf.ConsolePort)
	server.MaxConnNum = int(math.MaxInt32)
	server.PendingWriteNum = 100
	server.NewAgent = newAgent

	server.Start()
}

func Destroy() {
	if server != nil {
		server.Close()
	}
}

type Agent struct {
	conn   *network.TCPConn
	reader *bufio.Reader
}

func newAgent(conn *network.TCPConn) network.Agent {
	a := new(Agent)
	a.conn = conn
	a.reader = bufio.NewReader(conn)
	return a
}

func (a *Agent) Run() {
	for {
		if conf.ConsolePrompt != "" {
			a.conn.Write([]byte(conf.ConsolePrompt))
		}

		line, err := a.reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSuffix(line[:len(line)-1], "\r")

		args := strings.Fields(line)
		if len(args) == 0 {
			continue
		}
		if args[0] == "quit" {
			break
		}
		var c Command
		for _, _c := range commands {
			if _c.name() == args[0] {
				c = _c
				break
			}
		}
		if c == nil {
			a.conn.Write([]byte("command not found, try `help` for help\r\n"))
			continue
		}
		output := c.run(args[1:])
		if output != "" {
			a.conn.Write([]byte(output + "\r\n"))
		}
	}
}

func (a *Agent) OnClose() {}
