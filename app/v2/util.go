package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
)

// WaitSignal waits for a signal to exit the program.
func WaitSignal(ctx context.Context, errC <-chan error, hook func()) {
	signC := make(chan os.Signal, 1)
	signal.Notify(signC, os.Interrupt, os.Kill)
	select {
	case <-signC:
		log.Println("received interrupt signal")
		hook()
		os.Exit(0)
	case <-ctx.Done():
		log.Println("context done")
		hook()
		os.Exit(0)
	case err := <-errC:
		if err != nil {
			log.Printf("run error: %+v", err)
			hook()
			os.Exit(1)
		} else {
			log.Println("run successfully")
			hook()
			os.Exit(0)
		}
	}
}

// If returns a if cond is true, otherwise b.
func If[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}

// PrettyString prints the value in a pretty format.
func PrettyString(v any) string {
	if v == nil {
		return "nil"
	}

	if j, err := json.Marshal(v); err == nil {
		return string(j)
	}

	return fmt.Sprintf("%+v", v)
}
