package leaf

import (
	"github.com/name5566/leaf/log"
	"os"
	"os/signal"
)

type App interface {
	OnInit()
	OnDestroy()
}

type Config struct {
	LogLevel string
	LogPath  string
}

var c Config

func SetConfig(_c Config) {
	c = _c
}

func Run(app App) {
	if c.LogLevel != "" {
		logger, err := log.New(c.LogLevel, c.LogPath)
		if err != nil {
			panic(err)
		}
		log.Export(logger)
		defer logger.Close()
	}

	log.Release("Leaf starting up")
	app.OnInit()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	log.Release("Leaf closing down (signal: %v)", sig)
	app.OnDestroy()
}
