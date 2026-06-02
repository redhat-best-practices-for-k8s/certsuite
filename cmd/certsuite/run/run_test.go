// Copyright (C) 2024-2026 Red Hat, Inc.
package run

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("name", "default", "")
	cmd.Flags().Bool("verbose", false, "")
	return cmd
}

func TestFlagReaderGetString(t *testing.T) {
	t.Parallel()
	cmd := newTestCommand()
	f := &flagReader{cmd: cmd}

	var val string
	f.getString(&val, "name")

	require.NoError(t, f.err)
	assert.Equal(t, "default", val)
}

func TestFlagReaderGetStringExplicitValue(t *testing.T) {
	t.Parallel()
	cmd := newTestCommand()
	require.NoError(t, cmd.Flags().Set("name", "custom"))
	f := &flagReader{cmd: cmd}

	var val string
	f.getString(&val, "name")

	require.NoError(t, f.err)
	assert.Equal(t, "custom", val)
}

func TestFlagReaderGetBool(t *testing.T) {
	t.Parallel()
	cmd := newTestCommand()
	f := &flagReader{cmd: cmd}

	var val bool
	f.getBool(&val, "verbose")

	require.NoError(t, f.err)
	assert.False(t, val)
}

func TestFlagReaderGetStringUnregisteredFlag(t *testing.T) {
	t.Parallel()
	cmd := newTestCommand()
	f := &flagReader{cmd: cmd}

	var val string
	f.getString(&val, "nonexistent")

	require.Error(t, f.err)
	assert.Contains(t, f.err.Error(), `"nonexistent"`)
	assert.Empty(t, val)
}

func TestFlagReaderGetBoolUnregisteredFlag(t *testing.T) {
	t.Parallel()
	cmd := newTestCommand()
	f := &flagReader{cmd: cmd}

	var val bool
	f.getBool(&val, "nonexistent")

	require.Error(t, f.err)
	assert.Contains(t, f.err.Error(), `"nonexistent"`)
	assert.False(t, val)
}

func TestFlagReaderShortCircuitsAfterError(t *testing.T) {
	t.Parallel()
	cmd := newTestCommand()
	f := &flagReader{cmd: cmd}

	var bad string
	f.getString(&bad, "nonexistent")
	require.Error(t, f.err)
	firstErr := f.err

	var good string
	f.getString(&good, "name")

	assert.Empty(t, good, "should not have been populated after a prior error")
	assert.Equal(t, firstErr, f.err, "error should not have changed")
}

func TestFlagReaderGetBoolOnStringFlag(t *testing.T) {
	t.Parallel()
	cmd := newTestCommand()
	f := &flagReader{cmd: cmd}

	var val bool
	f.getBool(&val, "name")

	require.Error(t, f.err)
	assert.Contains(t, f.err.Error(), `"name"`)
}

func TestFlagReaderGetStringOnBoolFlag(t *testing.T) {
	t.Parallel()
	cmd := newTestCommand()
	f := &flagReader{cmd: cmd}

	var val string
	f.getString(&val, "verbose")

	require.Error(t, f.err)
	assert.Contains(t, f.err.Error(), `"verbose"`)
}
