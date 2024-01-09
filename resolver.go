package commander

import (
	"context"
	"fmt"
	"reflect"
)

var (
	resolveString   = ResolverFunc(func(arg string) (string, error) { return arg, nil })
	resolveContext  = ContextResolverFunc(func(ctx context.Context) (context.Context, error) { return ctx, nil })
	resolveAnything = new(fallbackResolver)

	contextType = reflect.TypeOf((*context.Context)(nil)).Elem()
	stringType  = reflect.TypeOf((*string)(nil)).Elem()
)

type Resolver interface {
	CanResolve(typ reflect.Type) bool
	RequiresArg(typ reflect.Type) bool
	Resolve(typ reflect.Type, ctx *ResolverContext) (reflect.Value, error)
}

type ResolverContext struct {
	context.Context
	Args  []string
	State map[Resolver]any
}

func (ctx *ResolverContext) NextArg() string {
	arg := ctx.Args[0]
	ctx.Args = ctx.Args[1:]
	return arg
}

func ResolverFunc[T any](resolve func(arg string) (T, error)) Resolver {
	var t T
	return &resolver{
		typ:    reflect.TypeOf(t),
		useArg: true,
		resolve: func(ctx *ResolverContext) (reflect.Value, error) {
			v, err := resolve(ctx.NextArg())
			return reflect.ValueOf(v), err
		},
	}
}

func ContextResolverFunc[T any](resolve func(ctx context.Context) (T, error)) Resolver {
	var t T
	return &resolver{
		typ:    reflect.TypeOf(t),
		useArg: false,
		resolve: func(ctx *ResolverContext) (reflect.Value, error) {
			v, err := resolve(ctx.Context)
			return reflect.ValueOf(v), err
		},
	}
}

type resolverBinding struct {
	typ      reflect.Type
	resolver Resolver
}

func (b *resolverBinding) requiresArg() bool {
	return b.resolver.RequiresArg(b.typ)
}

func (b *resolverBinding) resolve(ctx *ResolverContext) (reflect.Value, error) {
	return b.resolver.Resolve(b.typ, ctx)
}

func bindResolver(typ reflect.Type, resolver Resolver) resolverBinding {
	return resolverBinding{
		typ:      typ,
		resolver: resolver,
	}
}

type resolver struct {
	typ     reflect.Type
	useArg  bool
	resolve func(ctx *ResolverContext) (reflect.Value, error)
}

func (r *resolver) CanResolve(typ reflect.Type) bool {
	return r.typ.AssignableTo(typ)
}

func (r *resolver) RequiresArg(typ reflect.Type) bool {
	return r.useArg
}

func (r *resolver) Resolve(typ reflect.Type, ctx *ResolverContext) (reflect.Value, error) {
	return r.resolve(ctx)
}

type fallbackResolver struct{}

func (r *fallbackResolver) CanResolve(typ reflect.Type) bool {
	return true
}

func (r *fallbackResolver) RequiresArg(typ reflect.Type) bool {
	return true
}

func (r *fallbackResolver) Resolve(typ reflect.Type, ctx *ResolverContext) (v reflect.Value, err error) {
	arg := ctx.NextArg()
	defer func() {
		if r := recover(); r != nil {
			err = &ArgConversionError{
				Arg:        arg,
				TargetType: typ,
				Err:        fmt.Errorf("%v", r),
			}
		}
	}()
	result := reflect.New(typ)
	_, err = fmt.Sscan(arg, result.Interface())
	if err != nil {
		err = &ArgConversionError{
			Arg:        arg,
			TargetType: typ,
			Err:        err,
		}
	}
	return reflect.Indirect(result), err
}

func findResolver(typ reflect.Type, resolvers []Resolver) Resolver {
	for _, r := range resolvers {
		if r.CanResolve(typ) {
			return r
		}
	}
	switch typ {
	case stringType:
		return resolveString
	case contextType:
		return resolveContext
	default:
		return resolveAnything
	}
}
