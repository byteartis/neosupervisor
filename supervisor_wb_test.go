//nolint:testpackage // Test it's easier with white-box testing
package neosupervisor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestApplicationLaunch(t *testing.T) {
	t.Parallel()

	app := New()
	require.True(t, app.isNew())

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)

	go func() {
		errCh <- app.Launch(ctx)
	}()

	time.Sleep(500 * time.Millisecond)

	err := app.AddRunner(context.Background(), new(MockRunner))
	require.ErrorIs(t, err, ErrApplicationIsRunning)

	go func() {
		time.Sleep(1 * time.Second)
		cancel()
	}()

	require.False(t, app.isNew() || app.isStopped())
	err = app.Launch(ctx)
	require.NoError(t, err)

	err = <-errCh
	require.NoError(t, err)
	require.True(t, app.isStopped())

	err = app.Launch(ctx)
	require.ErrorIs(t, err, ErrApplicationAlreadyStopped)
}

func TestRegisterRunner(t *testing.T) {
	t.Parallel()

	t.Run("when launch returns an error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("some error")

		mockRunner := new(MockRunner)
		mockRunner.On("Run", mock.Anything).Return(expectedErr)

		app := New()

		// Register the runner
		err := app.AddRunner(context.Background(), mockRunner)
		require.NoError(t, err)

		err = app.Launch(context.Background())
		assert.Equal(t, expectedErr, err)
		mockRunner.AssertExpectations(t)
		assert.Len(t, mockRunner.Calls, 1)
	})

	t.Run("when launch is successful", func(t *testing.T) {
		t.Parallel()

		mockRunner := new(MockRunner)
		mockRunner.On("Run", mock.Anything).Return(nil)

		app := New()

		// Register the runner
		err := app.AddRunner(context.Background(), mockRunner, WithRunnerName("test"))
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())

		errCh := make(chan error, 1)

		go func() {
			errCh <- app.Launch(ctx)
		}()

		time.Sleep(500 * time.Millisecond)
		cancel()

		timer := time.NewTimer(3 * time.Second)

		for {
			select {
			case err := <-errCh:
				require.NoError(t, err, "app failed to launch")
			case <-timer.C:
				require.FailNow(t, "application took to long to stop")
			default:
				if app.isStopped() {
					mockRunner.AssertExpectations(t)
					return
				}
			}
		}
	})
}

func TestShutdown(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario        string
		gracefulTimeout int
		runnerTimeout   time.Duration
	}{
		{
			scenario:        "when graceful shutdown takes longer than runners to stop",
			gracefulTimeout: 2,
			runnerTimeout:   1 * time.Second,
		},
		{
			scenario:        "when graceful shutdown takes less than runners to stop",
			gracefulTimeout: 1,
			runnerTimeout:   2 * time.Second,
		},
	}

	for ti := range testCases {
		tc := testCases[ti]

		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			app := New(
				WithTimeout(tc.gracefulTimeout),
			)
			require.True(t, app.isNew())

			mockRunner := MockRunner{sleepDuration: tc.runnerTimeout}
			mockRunner.On("Run", mock.Anything).Return(nil)

			err := app.AddRunner(context.Background(), &mockRunner)
			require.NoError(t, err)

			ctx, cancel := context.WithCancel(context.Background())
			errCh := make(chan error, 1)

			go func() {
				errCh <- app.Launch(ctx)
			}()

			time.Sleep(500 * time.Millisecond)
			cancel()

			timer := time.NewTimer(3 * time.Second)

			for {
				select {
				case err := <-errCh:
					require.NoError(t, err, "app failed to launch")
				case <-timer.C:
					require.FailNow(t, "application took to long to stop")
				default:
					if app.isStopped() {
						mockRunner.AssertExpectations(t)
						return
					}
				}
			}
		})
	}
}

type MockRunner struct {
	mock.Mock
	sleepDuration time.Duration
}

func (m *MockRunner) Run(ctx context.Context) error {
	args := m.Called(ctx)
	err := args.Error(0)
	if err == nil {
		<-ctx.Done()
		if m.sleepDuration >= 0 {
			time.Sleep(m.sleepDuration)
		}
		return nil
	}
	return err
}
