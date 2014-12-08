package leaf

import (
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/module"
	"os"
	"os/signal"
)

type Config struct {
	LogLevel string
	LogPath  string
}

var c Config

func SetConfig(_c Config) {
	c = _c
}

func Run(mods ...module.Module) {
	// logger
	if c.LogLevel != "" {
		logger, err := log.New(c.LogLevel, c.LogPath)
		if err != nil {
			panic(err)
		}
		log.Export(logger)
		defer logger.Close()
	}

	log.Release("Leaf starting up")

	// module
	for i := 0; i < len(mods); i++ {
		module.Register(mods[i])
	}
	module.Init()

	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	log.Release("Leaf closing down (signal: %v)", sig)
	module.Destroy()
}
