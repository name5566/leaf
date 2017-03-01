package cluster

import (
	"github.com/name5566/leaf/network/json"
)

var (
	Processor = json.NewProcessor()
)

type S2S_NotifyServerName struct {
	ServerName	string
}

func handleNotifyServerName(args []interface{}) {
	msg := args[0].(*S2S_NotifyServerName)
	agent := args[1].(*Agent)
	agent.ServerName = msg.ServerName
	agents[agent.ServerName] = agent
}

func init() {
	Processor.SetHandler(&S2S_NotifyServerName{}, handleNotifyServerName)
}