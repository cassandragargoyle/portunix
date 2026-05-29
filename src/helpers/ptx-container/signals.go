/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"os"
	"os/signal"
	"syscall"
)

// registerSignalShutdown closes the supplied channel on SIGTERM/SIGINT so the
// lifecycle service main loop can break out of its select. Issue #027 calls
// for "Cleanup on process exit (SIGTERM/SIGINT handlers)"; this is the hook.
//
// We deliberately do NOT call os.Exit from here — the loop tears down its own
// resources (pid file removal, log line) on graceful return.
//
// stop is a bidirectional chan so the goroutine can both check for an existing
// close (to be safe against double-signal) and close it.
func registerSignalShutdown(stop chan struct{}) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-c
		logf("received signal %s, shutting down", sig)
		// Non-blocking close so a doubled signal doesn't panic on a closed chan.
		select {
		case <-stop:
		default:
			close(stop)
		}
	}()
}
