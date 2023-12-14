package commander

import (
	"fmt"
	"reflect"
)

var ErrNotFunction = fmt.Errorf("not a function")

type UnknownCommandError struct {
	Cmd string
}

func (e UnknownCommandError) Error() string {
	return "unknown command: " + e.Cmd
}

type ArgCountMismatchError struct {
	Want     int
	Got      int
	Variadic bool
}

func (e ArgCountMismatchError) Error() string {
	if e.Variadic {
		return fmt.Sprintf("expected at least %d argument(s), got %d", e.Want, e.Got)
	}
	return fmt.Sprintf("expected %d argument(s), got %d", e.Want, e.Got)
}

type ArgConversionError struct {
	Arg        string
	TargetType reflect.Type
	Err        error
}

func (e ArgConversionError) Error() string {
	return fmt.Sprintf("failed to convert arg %q to %s", e.Arg, e.TargetType)
}

func (e ArgConversionError) Unwrap() error {
	return e.Err
}

type CommandRuntimeError struct {
	Err error
}

func (e CommandRuntimeError) Error() string {
	return fmt.Sprintf("command runtime error: %v", e.Err)
}

func (e CommandRuntimeError) Unwrap() error {
	return e.Err
}
