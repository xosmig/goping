package goping

import (
	"flag"
	"fmt"
	"io"
)

type Params struct {
	Timeout  int
	Interval int
	Count    int
	Deadline int
	Url      string
}

const (
	DefaultTimeout  int = 2
	DefaultInterval int = 3
	DefaultCount    int = -1  // not limited
	DefaultDeadline int = -1  // not limited
)

type WrongNumberOfArguments struct {
	expected, actual int
}

func (err WrongNumberOfArguments) Error() string {
	return fmt.Sprintf("wrong number of arguments. Expected: %d, actual: %d", err.expected, err.actual)
}

func ParseCommandLine(args []string, errorOutput io.Writer) (Params, error) {
	commandLine := flag.NewFlagSet("goping", 0)
	commandLine.SetOutput(errorOutput)

	commandLine.Usage = func() {
		fmt.Fprintln(errorOutput, "Usage: goping [options] url")
		fmt.Fprintln(errorOutput, "Available options:")
		commandLine.PrintDefaults()
	}

	var result Params
	commandLine.IntVar(&result.Timeout, "timeout", DefaultTimeout, "Time to wait for a response in seconds.")
	commandLine.IntVar(&result.Interval, "interval", DefaultInterval, "Minimum interval between attempts in seconds.")
	commandLine.IntVar(&result.Count, "count", -1, "Stop  after  sending  Count  ECHO_REQUEST  packets.")
	commandLine.IntVar(&result.Deadline, "deadline", DefaultDeadline, "Specify a timeout, in seconds, "+
		"before ping exits regardless of how many packets have been sent or received.")

	if err := commandLine.Parse(args); err != nil {
		return Params{}, err
	}

	if commandLine.NArg() != 1 {
		commandLine.Usage()
		return Params{}, WrongNumberOfArguments{expected: 1, actual: commandLine.NArg()}
	}

	result.Url = commandLine.Arg(0)
	return result, nil
}
