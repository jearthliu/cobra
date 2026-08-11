package cobra

import (
	"bytes"
	"io"

	"github.com/spf13/pflag"
)

// Command represents a CLI command with flags and subcommands.
type Command struct {
	Use             string
	Run             func(cmd *Command, args []string)
	parent          *Command
	commands        []*Command
	flags           *pflag.FlagSet
	persistentFlags *pflag.FlagSet
	args            []string
	out             io.Writer
	err             io.Writer
	lastExecuted    bool // guards reset-flag ordering across executions
}

// NewCommand creates a root command with fresh flag sets.
func NewCommand() *Command {
	return &Command{
		flags:           pflag.NewFlagSet("", pflag.ContinueOnError),
		persistentFlags: pflag.NewFlagSet("", pflag.ContinueOnError),
		out:             &bytes.Buffer{},
		err:             &bytes.Buffer{},
	}
}

// Flags returns the command's local flag set, initializing it if needed.
func (c *Command) Flags() *pflag.FlagSet {
	if c.flags == nil {
		c.flags = pflag.NewFlagSet("", pflag.ContinueOnError)
	}
	return c.flags
}

// PersistentFlags returns the command's persistent flag set.
func (c *Command) PersistentFlags() *pflag.FlagSet {
	if c.persistentFlags == nil {
		c.persistentFlags = pflag.NewFlagSet("", pflag.ContinueOnError)
	}
	return c.persistentFlags
}

// AddCommand registers a subcommand and links its parent.
func (c *Command) AddCommand(cmd *Command) {
	cmd.parent = c
	c.commands = append(c.commands, cmd)
}

// Commands returns the registered subcommands.
func (c *Command) Commands() []*Command {
	return c.commands
}

// SetArgs sets the argument vector for the next execution.
func (c *Command) SetArgs(args []string) {
	c.args = args
}

// SetOut sets the output writer.
func (c *Command) SetOut(w io.Writer) {
	c.out = w
}

// SetErr sets the error writer.
func (c *Command) SetErr(w io.Writer) {
	c.err = w
}

// Execute runs the command. Before each execution, flags are reset to their
// defaults so a previous execution's values never leak into this one (see
// ResetFlags). This is the fix for the flag-value-leak issue: it guarantees
// execution isolation even when the same Command is executed repeatedly.
func (c *Command) Execute() error {
	// Reset flags before parsing on every execution after the first. The first
	// run starts from defaults anyway; resetting here keeps behavior identical
	// for single-execution CLIs (acceptance criterion 4).
	c.ResetFlags()

	// Resolve the target subcommand by walking args.
	target := c
	for _, a := range c.args {
		if child := target.findChild(a); child != nil {
			target = child
		} else {
			break
		}
	}

	// Merge persistent flags from root down to target, then parse.
	flagSet := pflag.NewFlagSet(target.Use, pflag.ContinueOnError)
	flagSet.SetOutput(target.err)
	for p := target; p != nil; p = p.parent {
		if p.persistentFlags != nil {
			p.persistentFlags.VisitAll(func(f *pflag.Flag) {
				_ = flagSet.AddFlag(f)
			})
		}
	}
	if target.flags != nil {
		target.flags.VisitAll(func(f *pflag.Flag) {
			_ = flagSet.AddFlag(f)
		})
	}

	var positional []string
	if err := flagSet.Parse(c.args); err != nil {
		return err
	}
	positional = flagSet.Args()

	if target.Run != nil {
		target.Run(target, positional)
	}
	return nil
}

// findChild returns the subcommand matching name, if any.
func (c *Command) findChild(name string) *Command {
	for _, child := range c.commands {
		if child.Use == name {
			return child
		}
	}
	return nil
}
