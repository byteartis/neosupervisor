//nolint:testpackage // Test it's easier with white-box testing
package neosupervisor

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		options []Option
		assert  func(*testing.T, Options)
	}{
		{
			name:    "no options provided",
			options: []Option{},
			assert: func(t *testing.T, o Options) {
				t.Helper()
				assert.Equal(t, time.Duration(0)*time.Second, o.timeout)
			},
		},
		{
			name:    "invalid timeout provided",
			options: []Option{WithTimeout(-10)},
			assert: func(t *testing.T, o Options) {
				t.Helper()
				assert.Equal(t, time.Duration(0)*time.Second, o.timeout)
			},
		},
		{
			name:    "valid timeout provided",
			options: []Option{WithTimeout(10)},
			assert: func(t *testing.T, o Options) {
				t.Helper()
				assert.Equal(t, time.Duration(10)*time.Second, o.timeout)
			},
		},
	}

	for ti := range tests {
		tc := tests[ti]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			options := Options{}

			for _, o := range tc.options {
				o(&options)
			}

			tc.assert(t, options)
		})
	}
}

func TestNewWithOptions(t *testing.T) {
	t.Parallel()

	app := New(WithTimeout(1))
	require.True(t, app.isNew())
	assert.Equal(t, 1*time.Second, app.opts.timeout)
}

type MockSubscription struct {
	mock.Mock
}

func (m *MockSubscription) Run(ctx context.Context) error {
	args := m.Called(ctx)
	err := args.Error(0)
	if err == nil {
		<-ctx.Done()
	}
	return err
}
