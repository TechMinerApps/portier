package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/TechMinerApps/portier/app"
)

func main() {
	app := app.NewPortier()
	app.Start()
	sigchan := make(chan os.Signal)
	signal.Notify(sigchan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGHUP,
		syscall.SIGUSR1,
		syscall.SIGUSR2,
		syscall.SIGSTOP)

	// Graceful Shutdown
	go func() {
		sig := <-sigchan
		app.Stop(sig)
	}()
	app.Wait()
}
