package app

import (
	"context"
	"fmt"
	syslog "log"
	"os"
	"os/signal"
)

func (a *application) run(c *Context, confPtr any, cancelAppContext context.CancelFunc) error {
	if a.prepare != nil {
		err := a.prepare(c, confPtr)
		if err != nil {
			return fmt.Errorf("prepare function error, err: %w", err)
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
	cleanupFunctions := make([]func(), 0)
	for _, initFunction := range a.initFunctions {
		cleanupFunc, err := initFunction(c)
		if err != nil {
			return fmt.Errorf("init function error, err: %w", err)
		}
		if cleanupFunc != nil {
			cleanupFunctions = append(cleanupFunctions, cleanupFunc)
		}
	}

	// Run the main logic
	targetRunFn := a.runFunction

	if err := targetRunFn(c, confPtr, waitFunc); err != nil {
		return fmt.Errorf("main function error, err: %w", err)
	}

	for _, cleanupFunction := range cleanupFunctions {
		cleanupFunction()
	}

	return nil
}

func WaitSignal(errC <-chan error, hook func()) {
	signC := make(chan os.Signal, 1)
	signal.Notify(signC, os.Interrupt, os.Kill)
	select {
	case <-signC:
		syslog.Println("received interrupt signal")
		hook()
		os.Exit(0)
	case err := <-errC:
		syslog.Printf("run error: %+v", err)
		hook()
		os.Exit(1)
	}
}
