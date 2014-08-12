package leaf

import (
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/service/gate"
	"os"
	"os/signal"
)

type Config struct {
	LogLevel       string
	LogPath        string
	GateAddr       string
	GateMaxConnNum int
	GateAgentMgr   gate.AgentMgr
}

func Run(cfg Config) {
	// log
	logger, err := log.New(cfg.LogLevel, cfg.LogPath)
	if err != nil {
		panic(err)
	}
	log.Export(logger)
	defer log.Close()

	log.Release("Leaf server starting up")

	// gate
	gate, err := gate.NewTcpGate(cfg.GateAddr, cfg.GateMaxConnNum, cfg.GateAgentMgr)
	if err != nil {
		log.Fatal("%v", err)
	}
	gate.Start()

	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	log.Release("Leaf server closing down (signal: %v)", sig)

	gate.Close()
}
