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
	TcpGateConfig  *gate.TcpGateConfig
	HttpGateConfig *gate.HttpGateConfig
}

func Run(config Config) {
	// log
	if config.LogLevel != "" {
		logger, err := log.New(config.LogLevel, config.LogPath)
		if err != nil {
			panic(err)
		}
		log.Export(logger)
		defer logger.Close()
	}

	log.Release("Leaf server starting up")

	// gate
	if config.TcpGateConfig != nil {
		gate, err := gate.NewTcpGate(config.TcpGateConfig)
		if err != nil {
			log.Fatal("%v", err)
		}
		gate.Start()
		defer gate.Close()
	} else if config.HttpGateConfig != nil {
		gate, err := gate.NewHttpGate(config.HttpGateConfig)
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
