package app

import (
	"context"
	"fmt"
)

func (a *application) run(c *Context, confPtr any, cancelAppContext context.CancelFunc) error {
	var (
		errC = make(chan error)
	)

	// Run init functions
	for _, initFn := range a.initFns {
		// Run the init function
		cleanFn, err := initFn(c, confPtr)
		if err != nil {
			return fmt.Errorf("init error, err: %w", err)
		}

		if cleanFn != nil {
			a.cleanFns = append(a.cleanFns, cleanFn)
		}
	}

	// Run the main logic
	go a.runFn(c, confPtr, errC)
	WaitSignal(c, errC, func() {
		cancelAppContext()
		a.cleanup()
	})

	return nil
}

// cleanup will run all cleanup functions
func (a *application) cleanup() {
	for _, cleanFn := range a.cleanFns {
		cleanFn()
	}
}
