package console

import (
	"fmt"
	"github.com/name5566/leaf/conf"
	"os"
	"path"
	"runtime/pprof"
	"time"
)

var commands = []Command{
	new(CommandHelp),
	new(CommandCPUProf),
}

type Command interface {
	// must goroutine safe
	name() string
	// must goroutine safe
	help() string
	// must goroutine safe
	run(arg []string) string
}

// help
type CommandHelp struct{}

func (c *CommandHelp) name() string {
	return "help"
}

func (c *CommandHelp) help() string {
	return "This help text"
}

func (c *CommandHelp) run(arg []string) string {
	output := "Commands:\r\n"
	for i, c := range commands {
		output += c.name() + " - " + c.help()
		if i < len(commands)-1 {
			output += "\r\n"
		}
	}

	return output
}

// cpuprof
type CommandCPUProf struct{}

func (c *CommandCPUProf) name() string {
	return "cpuprof"
}

func (c *CommandCPUProf) help() string {
	return "CPU profiling for the current process"
}

func (c *CommandCPUProf) usage() string {
	return "Usage: cpuprof start|stop"
}

func (c *CommandCPUProf) run(arg []string) string {
	if len(arg) == 0 {
		return c.usage()
	}

	switch arg[0] {
	case "start":
		now := time.Now()
		fn := path.Join(conf.ProfilePath,
			fmt.Sprintf("%d%02d%02d_%02d_%02d_%02d.prof",
				now.Year(),
				now.Month(),
				now.Day(),
				now.Hour(),
				now.Minute(),
				now.Second()))
		f, err := os.Create(fn)
		if err != nil {
			return err.Error()
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			f.Close()
			return err.Error()
		}
		return fn
	case "stop":
		pprof.StopCPUProfile()
		return ""
	default:
		return c.usage()
	}
}
