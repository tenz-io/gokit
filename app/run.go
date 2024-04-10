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
	for _, initFn := range a.initFs {
		// Run the init function
		cleanFn, err := initFn(c, confPtr)
		if err != nil {
			return fmt.Errorf("init error, err: %w", err)
		}

		if cleanFn != nil {
			a.cleanFs = append(a.cleanFs, cleanFn)
		}
	}

	// Run the command function
	if a.commandF != nil {
		a.commandF(c, confPtr, errC)
	}

	// Run the main logic
	if a.runF != nil {
		go a.runF(c, confPtr, errC)

		WaitSignal(c, errC, func() {
			cancelAppContext()
			a.cleanup()
		})
	} else {
		cancelAppContext()
		a.cleanup()
	}

	return nil
}

// cleanup will run all cleanup functions
func (a *application) cleanup() {
	for _, cleanFn := range a.cleanFs {
		cleanFn()
	}
}
