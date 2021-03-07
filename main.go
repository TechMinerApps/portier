package main

import (
	"os"
	"os/signal"

	"github.com/TechMinerApps/portier/app"
)

func main() {
	app := app.NewPortier()
	app.Start()
	app.Wait()
	sigchan := make(chan os.Signal)
	signal.Notify(sigchan)

	// Graceful Shutdown
	go func() {
		sig := <-sigchan
		app.Stop(sig)
	}()
}
