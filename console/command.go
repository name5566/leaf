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
	new(CommandProf),
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
	return "cpuprof writes runtime profiling data in the format expected by \r\n" +
		"the pprof visualization tool\r\n\r\n" +
		"Usage: cpuprof start|stop\r\n" +
		"  start - enables CPU profiling\r\n" +
		"  stop  - stops the current CPU profile"
}

func (c *CommandCPUProf) run(arg []string) string {
	if len(arg) == 0 {
		return c.usage()
	}

	switch arg[0] {
	case "start":
		fn := profileName() + ".cpuprof"
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

func profileName() string {
	now := time.Now()
	return path.Join(conf.ProfilePath,
		fmt.Sprintf("%d%02d%02d_%02d_%02d_%02d",
			now.Year(),
			now.Month(),
			now.Day(),
			now.Hour(),
			now.Minute(),
			now.Second()))
}

// prof
type CommandProf struct{}

func (c *CommandProf) name() string {
	return "prof"
}

func (c *CommandProf) help() string {
	return "Writes a pprof-formatted snapshot"
}

func (c *CommandProf) usage() string {
	return "prof writes runtime profiling data in the format expected by \r\n" +
		"the pprof visualization tool\r\n\r\n" +
		"Usage: prof goroutine|heap|thread|block\r\n" +
		"  goroutine - stack traces of all current goroutines\r\n" +
		"  heap      - a sampling of all heap allocations\r\n" +
		"  thread    - stack traces that led to the creation of new OS threads\r\n" +
		"  block     - stack traces that led to blocking on synchronization primitives"
}

func (c *CommandProf) run(arg []string) string {
	if len(arg) == 0 {
		return c.usage()
	}

	var (
		p  *pprof.Profile
		fn string
	)
	switch arg[0] {
	case "goroutine":
		p = pprof.Lookup("goroutine")
		fn = profileName() + ".gprof"
	case "heap":
		p = pprof.Lookup("heap")
		fn = profileName() + ".hprof"
	case "thread":
		p = pprof.Lookup("threadcreate")
		fn = profileName() + ".tprof"
	case "block":
		p = pprof.Lookup("block")
		fn = profileName() + ".bprof"
	default:
		return c.usage()
	}

	f, err := os.Create(fn)
	if err != nil {
		return err.Error()
	}
	defer f.Close()
	err = p.WriteTo(f, 0)
	if err != nil {
		return err.Error()
	}

	return fn
}
