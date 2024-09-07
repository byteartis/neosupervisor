package neosupervisor

import (
	"time"
)

// Options the application's options
type Options struct {
	timeout time.Duration
	runners []runnerWrapper
}

// Option is a type alias for a functional option that can be passed to the supervisor
// during its creation.
type Option func(o *Options)

// WithTimeout allows to set the timeout for the supervisor.
// The provided value must be greater than 0 and it will override the default timeout.
func WithTimeout(timeout int) Option {
	return func(o *Options) {
		if timeout > 0 {
			o.timeout = time.Duration(timeout) * time.Second
		}
	}
}
