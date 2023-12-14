package commander

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestCase struct {
	Name            string
	Callback        any
	Ctx             context.Context
	Args            []string
	ExpectedResults []any
	ExpectedError   error
}

var testCases = []TestCase{
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
		Name:          "invalid arg count #1",
		Callback:      testCtxTestVarFound,
		Args:          []string{"1"},
		ExpectedError: new(ArgCountMismatchError),
	},
}

func TestCommand(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			cmd, err := NewCommand(tc.Callback)
			assert.NoError(t, err)
			require.NotNil(t, cmd)

			if tc.Ctx == nil {
				tc.Ctx = context.Background()
			}
			results, err := cmd.Call(tc.Ctx, tc.Args)
			require.Equal(t, len(tc.ExpectedResults), len(results))
			if tc.ExpectedError != nil {
				assert.ErrorAs(t, err, &tc.ExpectedError)
			} else {
				assert.NoError(t, err)
			}
			for i, r := range results {
				assert.Equal(t, tc.ExpectedResults[i], r)
			}
		})
	}
	t.Run("not a function", func(t *testing.T) {
		cmd, err := NewCommand(1)
		assert.Nil(t, cmd)
		assert.ErrorIs(t, err, ErrNotFunction)
	})
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
