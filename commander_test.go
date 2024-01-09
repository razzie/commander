package commander

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommander(t *testing.T) {
	cmdr := NewCommander()
	require.NotNil(t, cmdr)

	testcmdCalled := false
	testcmd := func() string {
		testcmdCalled = true
		return "ok"
	}

	err := cmdr.RegisterCommand("testcmd", testcmd)
	assert.NoError(t, err)

	v, err := cmdr.Call(context.Background(), "testcmd", nil)
	assert.Equal(t, []any{"ok"}, v)
	assert.NoError(t, err)
	assert.True(t, testcmdCalled)

	cmdr.UnregisterCommand("testcmd")

	var errUnknownCommand *UnknownCommandError
	_, err = cmdr.Call(context.Background(), "testcmd", nil)
	assert.ErrorAs(t, err, &errUnknownCommand)
}
