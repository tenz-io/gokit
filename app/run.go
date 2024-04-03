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
	cleanFns := make([]func(), 0, len(a.initFs))
	for _, initFn := range a.initFs {
		// Run the init function
		cleanFn, err := initFn(c, confPtr)
		if err != nil {
			return fmt.Errorf("init error, err: %w", err)
		}

		if cleanFn != nil {
			cleanFns = append(cleanFns, cleanFn)
		}
	}

	// Run the main logic
	go a.runF(c, confPtr, errC)

	WaitSignal(c, errC, func() {
		cancelAppContext()
		for _, cleanFn := range cleanFns {
			cleanFn()
		}
	})

	return nil
}
