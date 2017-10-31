package arguments

import (
	"flag"
	"github.com/calebgray/golibs/version"
	"os"
)

var help = flag.Bool("h", Verbose, "Be helpful.")
var helpmore = flag.Bool("hh", Verbose, "Be more helpful.")

var Verbose = false
var verbose = flag.Bool("v", Verbose, "Be verbose.")

type command struct {
	help     string
	function func([]string) int
}

var _commands map[string]command

func AddCommand(cmd string, help string, function func([]string) int) {
	if _commands == nil {
		_commands = make(map[string]command)
	}

	_commands[cmd] = command{
		help:     help,
		function: function,
	}
}

func Run() int {
	flag.Parse()

	Verbose = *verbose
	if Verbose {
		println(os.Args[0], "-", version.Get())
	}

	if len(os.Args) == 1 || *help || *helpmore {
		flag.Usage()

		if _commands != nil {
			println()
			println("Available Commands:")
			for name, cmd := range _commands {
				println(" ", name, "\t", cmd.help)
			}
		}

		return 1
	}

	if _commands != nil {
		cmd := _commands[flag.Args()[0]]
		if cmd.function != nil {
			return cmd.function(flag.Args()[1:])
		}
	}

	return -1
}
