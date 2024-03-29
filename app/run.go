package app

import (
	"context"
	"fmt"
)

func (a *application) run(c *Context, confPtr any, cancelAppContext context.CancelFunc) error {
	for _, prepare := range a.preparations {
		// Run the prepare function
		if err := prepare(c, confPtr); err != nil {
			return fmt.Errorf("prepare error, err: %w", err)
		}
	}

	errC := make(chan error)
	go WaitSignal(errC, cancelAppContext)

	waitFunc := func(userErrC ...<-chan error) {
		for _, uErrC := range userErrC {
			go func(ch <-chan error) {
				errC <- <-ch
			}(uErrC)
		}

		<-c.Context.Done()
	}

	// Run init functions
	cleanFns := make([]func(), 0)
	for _, initFn := range a.initFunctions {
		// Run the init function
		cleanFn, err := initFn(c)
		if err != nil {
			return fmt.Errorf("init function error, err: %w", err)
		}

		if cleanFn != nil {
			cleanFns = append(cleanFns, cleanFn)
		}
	}

	// Run the main logic
	targetRunFn := a.runFunction

	if err := targetRunFn(c, confPtr, waitFunc); err != nil {
		return fmt.Errorf("main function error, err: %w", err)
	}

	for _, cleanupFunction := range cleanFns {
		cleanupFunction()
	}

	return nil
}
