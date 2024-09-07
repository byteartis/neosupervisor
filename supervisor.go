package neosupervisor

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

// defaultTimeout defines the default graceful shutdown timeout
const defaultTimeout = 25

// Possible service states
const (
	serviceNEW int32 = iota
	serviceRUNNING
	serviceSTOPPING
	serviceSTOPPED
)

// New creates a new supervisor with the provided options.
func New(opts ...Option) *Supervisor {
	// Init and apply options
	options := Options{
		runners: []runnerWrapper{},
		timeout: defaultTimeout * time.Second,
	}
	for _, o := range opts {
		o(&options)
	}

	// Initialize the service
	return &Supervisor{
		opts:  options,
		state: serviceNEW,
		mtx:   sync.Mutex{},
	}
}

// Supervisor defines the entity responsible for running the different types of components.
type Supervisor struct {
	opts  Options
	state int32
	mtx   sync.Mutex
}

// AddRunner registers a new runner in the supervisor. Optionally we can attach several options to it.
func (a *Supervisor) AddRunner(ctx context.Context, r Runner, opts ...RunnerOption) error {
	a.mtx.Lock()
	defer a.mtx.Unlock()
	switch a.state {
	case serviceNEW:
		a.opts.runners = append(a.opts.runners, wrapRunner(ctx, r, opts...))
		return nil
	default:
		return ErrApplicationIsRunning
	}
}

// Launch will start the supervisor and all its components.
// It receives a context that when cancelled will also stop all components. Runners will be stopped immediately.
//
// If any failure occurs on Launch it will return an error.
//
// When called multiple times, if the supervisor is stopped will return an error, ottherwise it will immediately return nil.
// The possible states are: NEW, RUNNING, STOPPING, STOPPED.
func (a *Supervisor) Launch(launchCtx context.Context) error {
	switch atomic.LoadInt32(&a.state) {
	case serviceSTOPPED:
		// If the service was already stopped we cannot start it again
		return ErrApplicationAlreadyStopped
	case serviceNEW:
		atomic.StoreInt32(&a.state, serviceRUNNING)
		log.Ctx(launchCtx).Debug().Msg("application starting")
	default:
		// Only run if service is new
		return nil
	}

	runningCtx, cancelRunningCtx := context.WithCancel(log.Ctx(launchCtx).WithContext(context.Background()))
	defer cancelRunningCtx()

	runnersGroup := sync.WaitGroup{}
	errCh := make(chan error, 1)
	a.launchRunners(runningCtx, &runnersGroup, errCh)

	log.Ctx(launchCtx).Debug().Msg("application running")

	// Block waiting for shutdown
	select {
	case err := <-errCh:
		return err
	case <-launchCtx.Done():
		atomic.StoreInt32(&a.state, serviceSTOPPING)

		// Start context with cancellation set for the defined timeout
		log.Ctx(launchCtx).Debug().Msgf("application stopping in %.0fs", a.timeout().Seconds())

		// Prepare for graceful shutdown
		gracefulShutdownCtx, cancelGracefulShutdownCtx := context.WithTimeout(context.Background(), a.timeout())
		defer cancelGracefulShutdownCtx()

		// Cancel the running context
		cancelRunningCtx()

		runnersCloseCh := make(chan struct{}, 1)

		// wait for all runners to stop
		go func() {
			runnersGroup.Wait()
			runnersCloseCh <- struct{}{}
			close(runnersCloseCh)
		}()

		select {
		case <-runnersCloseCh:
			log.Ctx(launchCtx).Debug().Msg("all runners stopped")
		case <-gracefulShutdownCtx.Done():
			log.Ctx(launchCtx).Debug().Msg("graceful shutdown ended, application is about to die")
		}

		atomic.StoreInt32(&a.state, serviceSTOPPED)

		log.Ctx(launchCtx).Debug().Msg("application stopped")

		return nil
	}
}

/**
 * Runners Behavior
 */

// launchRunners attempts to run all runners
func (a *Supervisor) launchRunners(ctx context.Context, wg *sync.WaitGroup, errCh chan error) {
	a.mtx.Lock()
	defer a.mtx.Unlock()
	for _, runner := range a.opts.runners {
		wg.Add(1)
		go func(ctx context.Context, r runnerWrapper, ch chan error) {
			defer wg.Done()
			// Add runner name to the logging context
			if r.name != "" {
				ctx = log.Ctx(ctx).With().Str("runner", r.name).Logger().WithContext(ctx)
			}

			err := r.Run(ctx)
			if err != nil {
				ch <- err
				return
			}
		}(ctx, runner, errCh)
	}
}

/**
 * Test helpers... Perhaps we can remove them later
 */

// timeout returns the defined timeout for the application's graceful shutdown.
func (a *Supervisor) timeout() time.Duration {
	return a.opts.timeout
}

// isNew returns true if the service is in NEW state
func (a *Supervisor) isNew() bool {
	return atomic.LoadInt32(&a.state) == serviceNEW
}

// isStopped returns true if the service is in STOPPED state
func (a *Supervisor) isStopped() bool {
	return atomic.LoadInt32(&a.state) == serviceSTOPPED
}
