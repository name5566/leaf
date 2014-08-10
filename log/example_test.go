package log_test

import (
	"github.com/name5566/leaf/log"
)

func Example() {
	logger, err := log.New("release", "")
	if err != nil {
		return
	}

	logger.Debug("will not print")
	logger.Release("My name is %v", "Leaf")

	log.Export(logger)

	log.Debug("will not print")
	log.Release("123")
	log.Error("456")
	log.Fatal("789")
	log.Close()
}
