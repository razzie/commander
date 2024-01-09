package commander

import (
	"context"
	"fmt"
	"reflect"
)

type Command struct {
	fn                   reflect.Value
	fnType               reflect.Type
	numNonCtxInputs      int
	numOutputs           int
	inputHandlers        []ResolverBinding
	isVariadic           bool
	variadicInputHandler ResolverBinding
}

func NewCommand(callback any, resolvers ...Resolver) (*Command, error) {
	fn := reflect.ValueOf(callback)
	if fn.Kind() != reflect.Func {
		return nil, ErrNotFunction
	}

	fnType := fn.Type()
	isVariadic := fnType.IsVariadic()
	numInputs := fnType.NumIn()
	numOutputs := fnType.NumOut()

	numNonCtxInputs := 0
	inputHandlers := make([]ResolverBinding, numInputs)
	for i := range inputHandlers {
		inputType := fnType.In(i)
		if isVariadic && i == numInputs-1 {
			inputType = inputType.Elem()
		}
		resolver := findResolver(inputType, resolvers)
		if resolver.RequiresArg(inputType) {
			numNonCtxInputs++
		}
		inputHandlers[i] = BindResolver(inputType, resolver)
	}
	var variadicInputHandler ResolverBinding
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

func (cmd *Command) Call(ctx context.Context, args []string) (outputs []any, err error) {
	if err := cmd.checkArgs(args); err != nil {
		return nil, err
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	rctx := &ResolverContext{
		Context: ctx,
		Args:    args,
	}

	var inputs []reflect.Value
	for _, h := range cmd.inputHandlers {
		in, err := h(rctx)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, in)
	}
	for len(rctx.Args) > 0 {
		in, err := cmd.variadicInputHandler(rctx)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, in)
	}

	if cmd.numOutputs > 0 {
		outputs = make([]any, 0, cmd.numOutputs)
		for _, out := range cmd.fn.Call(inputs) {
			outputs = append(outputs, out.Interface())
		}
		if e, ok := outputs[len(outputs)-1].(error); ok {
			err = &CommandRuntimeError{Err: e}
		}
	}

	return
}
