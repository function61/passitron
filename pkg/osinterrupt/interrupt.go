package osinterrupt

import (
	"os"
	"os/signal"
	"syscall"
)

func WaitForIntOrTerm() os.Signal {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	return <-ch
}
