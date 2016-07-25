package console

import (
	"fmt"
	"github.com/name5566/leaf/chanrpc"
	"github.com/name5566/leaf/conf"
	"github.com/name5566/leaf/log"
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
	run(args []string) string
}

type ExternalCommand struct {
	_name  string
	_help  string
	server *chanrpc.Server
}

func (c *ExternalCommand) name() string {
	return c._name
}

func (c *ExternalCommand) help() string {
	return c._help
}

func (c *ExternalCommand) run(_args []string) string {
	args := make([]interface{}, len(_args))
	for i, v := range _args {
		args[i] = v
	}

	ret, err := c.server.Call1(c._name, args...)
	if err != nil {
		return err.Error()
	}
	output, ok := ret.(string)
	if !ok {
		return "invalid output type"
	}

	return output
}

// you must call the function before calling console.Init
// goroutine not safe
func Register(name string, help string, f interface{}, server *chanrpc.Server) {
	for _, c := range commands {
		if c.name() == name {
			log.Fatal("command %v is already registered", name)
		}
	}

	server.Register(name, f)

	c := new(ExternalCommand)
	c._name = name
	c._help = help
	c.server = server
	commands = append(commands, c)
}

// help
type CommandHelp struct{}

func (c *CommandHelp) name() string {
	return "help"
}

func (c *CommandHelp) help() string {
	return "this help text"
}

func (c *CommandHelp) run([]string) string {
	output := "Commands:\r\n"
	for _, c := range commands {
		output += c.name() + " - " + c.help() + "\r\n"
	}
	output += "quit - exit console"

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

func (c *CommandCPUProf) run(args []string) string {
	if len(args) == 0 {
		return c.usage()
	}

	switch args[0] {
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
	return "writes a pprof-formatted snapshot"
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

func (c *CommandProf) run(args []string) string {
	if len(args) == 0 {
		return c.usage()
	}

	var (
		p  *pprof.Profile
		fn string
	)
	switch args[0] {
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
