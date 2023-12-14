package commander

import (
	"context"
	"fmt"
	"reflect"
)

var (
	contextType = reflect.TypeOf((*context.Context)(nil)).Elem()
	stringType  = reflect.TypeOf((*string)(nil)).Elem()
)

type Command struct {
	fn                   reflect.Value
	fnType               reflect.Type
	numNonCtxInputs      int
	numOutputs           int
	inputHandlers        []commandInputHandler
	isVariadic           bool
	variadicInputHandler commandInputHandler
}

func NewCommand(callback any) (*Command, error) {
	fn := reflect.ValueOf(callback)
	if fn.Kind() != reflect.Func {
		return nil, ErrNotFunction
	}

	fnType := fn.Type()
	isVariadic := fnType.IsVariadic()
	numInputs := fnType.NumIn()
	numOutputs := fnType.NumOut()

	numNonCtxInputs := numInputs
	inputHandlers := make([]commandInputHandler, numInputs)
	for i := range inputHandlers {
		inputType := fnType.In(i)
		if isVariadic && i == numInputs-1 {
			inputType = inputType.Elem()
		}
		switch inputType {
		case contextType:
			numNonCtxInputs--
			inputHandlers[i] = cmdHandleCtx
		case stringType:
			inputHandlers[i] = cmdHandleStringArg
		default:
			inputHandlers[i] = cmdHandleAnyArg(inputType)
		}
	}
	var variadicInputHandler commandInputHandler
	if isVariadic {
		variadicInputHandler = inputHandlers[len(inputHandlers)-1]
		inputHandlers = inputHandlers[:len(inputHandlers)-1]
	}

	return &Command{
		fn:                   fn,
		fnType:               fnType,
		numNonCtxInputs:      numNonCtxInputs,
		numOutputs:           numOutputs,
		inputHandlers:        inputHandlers,
		isVariadic:           isVariadic,
		variadicInputHandler: variadicInputHandler,
	}, nil
}

func (cmd *Command) checkArgs(args []string) error {
	if cmd.isVariadic && len(args) < cmd.numNonCtxInputs-1 {
		return &ArgCountMismatchError{
			Want:     cmd.numNonCtxInputs - 1,
			Got:      len(args),
			Variadic: true,
		}
	} else if !cmd.isVariadic && len(args) != cmd.numNonCtxInputs {
		return &ArgCountMismatchError{
			Want: cmd.numNonCtxInputs,
			Got:  len(args),
		}
	}
	return nil
}

func (cmd *Command) Call(ctx context.Context, args []string) ([]any, error) {
	if err := cmd.checkArgs(args); err != nil {
		return nil, err
	}
	cc := &commandCall{
		cmd:  cmd,
		ctx:  ctx,
		args: args,
	}
	return cc.run()
}

type commandCall struct {
	cmd  *Command
	ctx  context.Context
	args []string
}

func (cc *commandCall) popArg() string {
	arg := cc.args[0]
	cc.args = cc.args[1:]
	return arg
}

func (cc *commandCall) run() (outputs []any, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	var inputs []reflect.Value
	for _, h := range cc.cmd.inputHandlers {
		in, err := h(cc)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, in)
	}
	for len(cc.args) > 0 {
		in, err := cc.cmd.variadicInputHandler(cc)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, in)
	}

	if cc.cmd.numOutputs > 0 {
		outputs = make([]any, 0, cc.cmd.numOutputs)
		for _, out := range cc.cmd.fn.Call(inputs) {
			outputs = append(outputs, out.Interface())
		}
		if e, ok := outputs[len(outputs)-1].(error); ok {
			err = &CommandRuntimeError{Err: e}
		}
	}

	return
}

type commandInputHandler func(*commandCall) (reflect.Value, error)

func cmdHandleCtx(cc *commandCall) (reflect.Value, error) {
	return reflect.ValueOf(cc.ctx), nil
}

func cmdHandleStringArg(cc *commandCall) (reflect.Value, error) {
	return reflect.ValueOf(cc.popArg()), nil
}

func cmdHandleAnyArg(targetType reflect.Type) commandInputHandler {
	return func(cc *commandCall) (v reflect.Value, err error) {
		arg := cc.popArg()
		defer func() {
			if r := recover(); r != nil {
				err = &ArgConversionError{
					Arg:        arg,
					TargetType: targetType,
					Err:        fmt.Errorf("%v", r),
				}
			}
		}()
		result := reflect.New(targetType)
		_, err = fmt.Sscan(arg, result.Interface())
		if err != nil {
			err = &ArgConversionError{
				Arg:        arg,
				TargetType: targetType,
				Err:        err,
			}
		}
		return reflect.Indirect(result), err
	}
}
