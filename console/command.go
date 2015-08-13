package console

var commands = []Command{
	new(CommandHelp),
}

type Command interface {
	// must goroutine safe
	Name() string
	// must goroutine safe
	Help() string
	// must goroutine safe
	Run(arg []string) string
}

// help
type CommandHelp struct{}

func (c *CommandHelp) Name() string {
	return "help"
}

func (c *CommandHelp) Help() string {
	return "This help text"
}

func (c *CommandHelp) Run(arg []string) string {
	output := "Commands:\r\n"
	for _, c := range commands {
		output += c.Name() + " - " + c.Help()
	}
	return output
}
