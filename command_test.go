package commander

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommand(t *testing.T) {
	type testcase struct {
		Name            string
		Callback        any
		CallbackError   error
		Ctx             context.Context
		Args            []string
		ExpectedResults []any
		ExpectedError   error
	}

	tests := []testcase{
		{
			Name:          "not a function",
			Callback:      1,
			CallbackError: new(NotFunctionError),
		},
		{
			Name:            "string arg",
			Callback:        strconv.Atoi,
			Args:            []string{"1"},
			ExpectedResults: []any{1, nil},
		},
		{
			Name:            "int arg",
			Callback:        strconv.Itoa,
			Args:            []string{"1"},
			ExpectedResults: []any{"1"},
		},
		{
			Name:            "context",
			Callback:        testCtxTestVarFound,
			Ctx:             context.WithValue(context.Background(), ctxTestVar, 1),
			ExpectedResults: []any{true},
		},
		{
			Name:            "variadic",
			Callback:        add,
			Args:            []string{"1", "2", "3"},
			ExpectedResults: []any{6},
		},
		{
			Name:          "arg conversion fail",
			Callback:      strconv.Itoa,
			Args:          []string{"a"},
			ExpectedError: new(ArgConversionError),
		},
		{
			Name:          "invalid arg count #1",
			Callback:      concat,
			Args:          []string{"1"},
			ExpectedError: new(ArgCountMismatchError),
		},
		{
			Name:          "invalid arg count #2",
			Callback:      testCtxTestVarFound,
			Args:          []string{"1"},
			ExpectedError: new(ArgCountMismatchError),
		},
		{
			Name:          "runtime error #1",
			Callback:      returnError,
			ExpectedError: new(CommandRuntimeError),
		},
		{
			Name:          "runtime error #2",
			Callback:      returnError,
			ExpectedError: new(customError),
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			cmd, err := NewCommand(tc.Callback)
			if tc.CallbackError != nil {
				assert.Nil(t, cmd)
				assert.ErrorAs(t, err, &tc.CallbackError)
				return
			}
			assert.NoError(t, err)
			require.NotNil(t, cmd)

			if tc.Ctx == nil {
				tc.Ctx = context.Background()
			}
			results, err := cmd.Call(tc.Ctx, tc.Args)
			if tc.ExpectedError != nil {
				assert.ErrorAs(t, err, &tc.ExpectedError)
			} else {
				assert.NoError(t, err)
				require.Equal(t, len(tc.ExpectedResults), len(results))
				for i, r := range results {
					assert.Equal(t, tc.ExpectedResults[i], r)
				}
			}
		})
	}
}

type ctxKey string

var ctxTestVar ctxKey = "ctxTestVar"

func testCtxTestVarFound(ctx context.Context) bool {
	return ctx.Value(ctxTestVar) != nil
}

func add(nums ...int) (sum int) {
	for _, n := range nums {
		sum += n
	}
	return
}

func concat(a, b string) string {
	return a + b
}

type customError struct{}

func (customError) Error() string {
	return "a custom error for testing"
}

func returnError() error {
	return &customError{}
}
