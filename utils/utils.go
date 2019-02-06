package utils

import (
	"fmt"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// TODO: Could get conflicts. Maybe closure is the answer?
var LogPgm string = "undefined"

func Log(format string, args ...interface{}) {
	log.Printf(LogPgm+": "+format, args...)
}

func FailOnError(err error, msg string) {
	if err != nil {
		Log(msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

// Get environment variable
func Getenv(env string, def string) string {
	s := os.Getenv(env)
	if s == "" {
		return def
	} else {
		return s
	}
}

func ContextWithSigterm(ctx context.Context) (context.Context, func()) {
	ctx, cncl := context.WithCancel(ctx)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	cancel := func() {
		Log("shutdown: preparing to shut down")
		signal.Stop(c)
		cncl()
		time.Sleep(time.Second * 1)
		Log("shutdown: shutdown complete, goodbye!")
	}

	go func() {
		select {
		case <-c:
			cncl()
		case <-ctx.Done():
		}
	}()

	return ctx, cancel
}
