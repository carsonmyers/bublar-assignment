package command

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/carsonmyers/bublar-assignment/logger"
	"go.uber.org/zap"
)

var log = logger.GetLogger()

// Command - a client command
type Command struct {
	name        string
	description string
	flags       *flag.FlagSet
	help        bool
	args        []string
	main        func(*Command) error
	next        []string
	commands    []*Command
}

// New - create a new client command
func New(name, description string, flags *flag.FlagSet, main func(*Command) error) *Command {
	cmd := &Command{
		name:        name,
		description: description,
		flags:       flags,
		args:        nil,
		main:        main,
		next:        nil,
		commands:    make([]*Command, 0),
	}

	if name == "help" {
		return cmd
	}

	cmd.AddCommand(helpCommand(cmd))

	if cmd.flags == nil {
		cmd.flags = flag.NewFlagSet(name, flag.ExitOnError)
	}

	cmd.flags.BoolVar(&cmd.help, "h", false, "")
	cmd.flags.BoolVar(&cmd.help, "help", false, "Display this help message")

	return cmd
}

// Name - get the command name
func (c *Command) Name() string {
	return c.name
}

// Init - initialize the command tree
func Init(name, description string, flags *flag.FlagSet, main func(*Command) error) *Command {
	command := New(name, description, flags, main)
	command.args, command.next = splitArgs(os.Args[1:])
	return command
}

// AddCommand - add a subcommand to the tree
func (c *Command) AddCommand(command *Command) {
	c.commands = append(c.commands, command)
}

// Execute - run the command
func (c *Command) Execute() error {
	log.Debug("Execute command", zap.String("command", c.name), zap.String("arguments", fmt.Sprintf("%v", c.args)))

	if c.flags != nil {
		c.flags.Parse(c.args)
	}

	if c.help {
		fmt.Print(c.Help())
		return nil
	}

	if err := c.main(c); err != nil {
		return fmt.Errorf("%s: %s", c.name, err)
	}

	return nil
}

// Help - get the help message
func (c *Command) Help() string {
	builder := strings.Builder{}
	builder.WriteString(c.name)
	builder.WriteString("\n\t")
	builder.WriteString(c.description)
	builder.WriteString("\n\n")

	if len(c.commands) > 0 {
		builder.WriteString("Subcommands:\n")
		for _, cmd := range c.commands {
			builder.WriteByte('\t')
			builder.WriteString(cmd.name)
			builder.WriteByte('\t')
			builder.WriteString(cmd.description)
			builder.WriteByte('\n')
		}
	}

	if c.flags != nil {
		builder.WriteString("\nFlags:\n")
		c.flags.VisitAll(func(f *flag.Flag) {
			if len(f.Name) == 1 {
				return
			}

			builder.WriteString("\t-")

			other := c.flags.Lookup(string(f.Name[0]))
			if other != nil && other.Usage == f.Usage {
				builder.WriteString(other.Name)
				builder.WriteString(" -")
			}

			builder.WriteString(f.Name)
			builder.WriteByte('\n')

			if len(f.Usage) > 0 {
				builder.WriteByte('\t')
				builder.WriteByte('\t')
				builder.WriteString(f.Usage)

				if len(f.DefValue) > 0 {
					builder.WriteString(" (default: ")
					builder.WriteString(f.DefValue)
					builder.WriteString(")\n")
				} else {
					builder.WriteByte('\n')
				}
			} else {
				builder.WriteByte('\n')
			}
		})
	}

	return builder.String()
}

// Args - get the command args
func (c *Command) Args() []string {
	return c.args
}

// Next - get the next command
func (c *Command) Next() (*Command, error) {
	return c.nextFrom(c.next)
}

func (c *Command) nextFrom(args []string) (*Command, error) {
	if len(args) == 0 {
		return nil, nil
	}

	for _, command := range c.commands {
		if command.name == args[0] {
			command.args, command.next = splitArgs(args[1:])
			return command, nil
		}
	}

	return nil, fmt.Errorf("Unknown command: %s", c.next[0])
}

func splitArgs(args []string) (this []string, next []string) {
	if len(args) > 0 {
		var seen bool
		for i, arg := range args {
			if arg[0] == '-' {
				seen = true
			} else if seen == false {
				this = args[:i]
				next = args[i:]
				return
			}

			if arg == "--" {
				this = args[:i]
				next = args[i:]
				return
			}
		}
	}

	return args, make([]string, 0)
}

func helpCommand(target *Command) *Command {
	return New("help", "Display this help message", nil, func(cmd *Command) error {
		next, err := target.nextFrom(cmd.next)
		if err != nil {
			return err
		}

		if next != nil {
			fmt.Print(next.Help())
		} else {
			fmt.Print(target.Help())
		}

		return nil
	})
}
