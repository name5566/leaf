package leaf

import (
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/server"
	"os"
	"os/signal"
)

type Config struct {
	LogLevel string
	LogPath  string
	Servers  []server.Server
}

func Run(config *Config) {
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

	// servers
	for _, s := range config.Servers {
		s.Start()
		defer s.Close()
	}

	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	log.Release("Leaf server closing down (signal: %v)", sig)
}
