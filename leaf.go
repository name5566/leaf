package leaf

import (
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/server"
	"os"
	"os/signal"
)

var servers []server.Server

func RegServer(server server.Server) {
	servers = append(servers, server)
}

func Run(logLevel string, logPath string) {
	// log
	if logLevel != "" {
		logger, err := log.New(logLevel, logPath)
		if err != nil {
			panic(err)
		}
		log.Export(logger)
		defer logger.Close()
	}

	log.Release("Leaf starting up")

	// servers
	if len(servers) == 0 {
		log.Fatal("server not found (call leaf.RegServer first)")
	}
	for _, s := range servers {
		s.Start()
		defer s.Close()
	}

	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	log.Release("Leaf closing down (signal: %v)", sig)
}
