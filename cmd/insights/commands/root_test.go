package commands

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ubuntu/ubuntu-insights/internal/constants"
)

func TestSetVerbosity(t *testing.T) {
	testCases := []struct {
		name    string
		pattern []bool
	}{
		{
			name:    "true",
			pattern: []bool{true},
		},
		{
			name:    "false",
			pattern: []bool{false},
		},
		{
			name:    "true false",
			pattern: []bool{true, false},
		},
		{
			name:    "false true",
			pattern: []bool{false, true},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, p := range tc.pattern {
				setVerbosity(p)

				if p {
					assert.True(t, slog.Default().Enabled(context.Background(), slog.LevelDebug))
				} else {
					assert.True(t, slog.Default().Enabled(context.Background(), constants.DefaultLogLevel))
				}
			}
		})
	}
}

func TestUsageError(t *testing.T) {
	app, err := New()
	require.NoError(t, err)

	// Test when SilenceUsage is true
	app.cmd.SilenceUsage = true
	assert.False(t, app.UsageError())

	// Test when SilenceUsage is false
	app.cmd.SilenceUsage = false
	assert.True(t, app.UsageError())
}
