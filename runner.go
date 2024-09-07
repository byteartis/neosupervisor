package neosupervisor

import (
	"context"
	"regexp"

	"github.com/rs/zerolog/log"
)

const runnerNameRgxStr = `^[a-z][a-z_]*$`

var runnerNameRgx = regexp.MustCompile(runnerNameRgxStr)

// Runner is the interface that must be implemented by all background running processes.
// The method must run while the received context is not cancelled.
// Returning an error will stop the entire supervisor and cause de application to exit with code 1.
type Runner interface {
	Run(context.Context) error
}

type runnerWrapper struct {
	Runner
	name string
}

func wrapRunner(ctx context.Context, r Runner, opts ...RunnerOption) runnerWrapper {
	options := runnerOptions{}
	for _, o := range opts {
		o(&options)
	}

	rw := runnerWrapper{
		Runner: r,
	}

	if options.name != "" && !runnerNameRgx.MatchString(options.name) {
		log.Ctx(ctx).
			Warn().
			Str("expected_regex", runnerNameRgxStr).
			Msgf("runner name '%s' is not valid", options.name)
	} else {
		rw.name = options.name
	}

	return rw
}

type runnerOptions struct {
	name string
}

// RunnerOption defines the options that can be passed to the supervisor when adding a new runner
type RunnerOption func(*runnerOptions)

// WithRunnerName defines the name of the runner in the options.
func WithRunnerName(n string) RunnerOption {
	return func(o *runnerOptions) {
		o.name = n
	}
}
