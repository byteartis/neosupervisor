package neosupervisor

import "errors"

var (
	// ErrApplicationIsRunning for when application is already running
	ErrApplicationIsRunning = errors.New("application is already running")
	// ErrApplicationAlreadyStopped for when the application is already stopped
	ErrApplicationAlreadyStopped = errors.New("application is stopped and cannot be started again")
)
