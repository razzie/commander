package commander

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	customType = reflect.TypeOf((*CustomType)(nil)).Elem()
	intType    = reflect.TypeOf((*int)(nil)).Elem()
)

func TestResolver(t *testing.T) {
	r1 := ResolverFunc[CustomType](func(arg string) (CustomType, error) {
		return CustomType{Val: arg}, nil
	})
	require.NotNil(t, r1)
	assert.True(t, r1.CanResolve(customType))
	assert.False(t, r1.CanResolve(intType))
	assert.True(t, r1.RequiresArg(customType))

	v1, err := r1.Resolve(customType, newContext(context.Background(), "arg1"))
	require.NotNil(t, v1)
	assert.NoError(t, err)
	assert.Equal(t, v1.Interface().(CustomType).Val, "arg1")

	r2 := ContextResolverFunc[CustomType](func(ctx context.Context) (CustomType, error) {
		arg := ctx.Value(ctxTestVar).(string)
		return CustomType{Val: arg}, nil
	})
	require.NotNil(t, r2)
	assert.True(t, r2.CanResolve(customType))
	assert.False(t, r2.CanResolve(intType))
	assert.False(t, r2.RequiresArg(customType))

	v2, err := r2.Resolve(customType, newContext(context.WithValue(context.Background(), ctxTestVar, "arg2")))
	require.NotNil(t, v2)
	assert.NoError(t, err)
	assert.Equal(t, v2.Interface().(CustomType).Val, "arg2")
}

func newContext(ctx context.Context, args ...string) *ResolverContext {
	return &ResolverContext{
		Context: ctx,
		Args:    args,
	}
}

type CustomType struct {
	Val string
}
