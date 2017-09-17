package goping

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

type Params struct {
	Timeout  time.Duration
	Count    int
	Deadline time.Duration
	Url      string
}

const (
	DefaultTimeout  time.Duration = time.Second * 2
	DefaultCount    int           = -1 // not limited
	DefaultDeadline time.Duration = -1 // not limited
)

type Cancelled struct {
	reason string
}

func (err Cancelled) Error() string {
	return "the call is canceled. Reason: " + err.reason
}

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
		fmt.Fprintf(errorOutput, "Usage: %s [flags] Url\n", os.Args[0])
		fmt.Fprintln(errorOutput, "Possible flags:")
		commandLine.PrintDefaults()
	}

	var result Params
	commandLine.DurationVar(&result.Timeout, "timeout", time.Second*2, "Time to wait for a response. "+
		"Include time units. Example: 1h20m, 1m, 1s, 100ms, 100ns.")
	commandLine.IntVar(&result.Count, "count", -1, "Stop  after  sending  Count  ECHO_REQUEST  packets.")
	commandLine.DurationVar(&result.Deadline, "deadline", -1, "Specify a timeout before ping exits regardless of how "+
		"many packets have been sent or received. Include units. Example: 1h20m, 1m, 1s, 100ms, 100ns.")
	helpPtr := commandLine.Bool("help", false, "Show this message.")

	if err := commandLine.Parse(args); err != nil {
		return Params{}, err
	}

	if *helpPtr {
		return Params{}, Cancelled{"help is called"}
	}

	if commandLine.NArg() != 1 {
		return Params{}, WrongNumberOfArguments{expected: 1, actual: commandLine.NArg()}
	}

	result.Url = commandLine.Arg(0)
	return result, nil
}
