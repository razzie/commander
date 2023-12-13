package commander

import (
	"context"
	"fmt"
)

type Commander struct {
	cmds map[string]Command
}

func NewCommander() *Commander {
	return &Commander{
		cmds: make(map[string]Command),
	}
}

func (cmdr *Commander) RegisterCommand(cmd string, callback any) error {
	c, err := NewCommand(callback)
	if err != nil {
		return err
	}
	cmdr.cmds[cmd] = *c
	return nil
}

func (cmdr *Commander) UnregisterCommand(cmd string) {
	delete(cmdr.cmds, cmd)
}

func (cmdr *Commander) Call(ctx context.Context, cmd string, args []string) ([]any, error) {
	c, ok := cmdr.cmds[cmd]
	if !ok {
		return nil, fmt.Errorf("unknown command: %s", cmd)
	}
	return c.Call(ctx, args)
}
