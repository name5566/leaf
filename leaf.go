package leaf

import (
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/service/gate"
	"os"
	"os/signal"
)

type Cfg struct {
	LogLevel    string
	LogPath     string
	tcpGateCfg  *gate.TcpGateCfg
	httpGateCfg *gate.HttpGateCfg
}

func Run(cfg Cfg) {
	// log
	if cfg.LogLevel != "" {
		logger, err := log.New(cfg.LogLevel, cfg.LogPath)
		if err != nil {
			panic(err)
		}
		log.Export(logger)
		defer logger.Close()
	}

	log.Release("Leaf server starting up")

	// gate
	if cfg.tcpGateCfg != nil {
		gate, err := gate.NewTcpGate(cfg.tcpGateCfg)
		if err != nil {
			log.Fatal("%v", err)
		}
		gate.Start()
		defer gate.Close()
	} else if cfg.httpGateCfg != nil {
		gate, err := gate.NewHttpGate(cfg.httpGateCfg)
		if err != nil {
			log.Fatal("%v", err)
		}
		gate.Start()
		defer gate.Close()
	} else {
		log.Fatal("gate config not found")
	}

	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	log.Release("Leaf server closing down (signal: %v)", sig)
}
