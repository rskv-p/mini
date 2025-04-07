package core

import (
	"sync"

	"github.com/rskv-p/mini/pkg/x_log"
)

// asyncCallbacksHandler manages async callback queue.
type asyncCallbacksHandler struct {
	cbQueue chan func() // Queue of callbacks
	closed  bool        // Indicates if handler is closed
	mu      sync.Mutex  // Mutex for synchronization
}

// run starts executing callbacks from the queue.
func (ac *asyncCallbacksHandler) run() {
	// Log callback loop start using global logger
	x_log.Debug().Msg("async callback loop started")

	for fn := range ac.cbQueue { // Process each callback in queue
		if fn == nil {
			// Log if nil callback received
			x_log.Warn().Msg("nil callback received")
			continue
		}
		fn() // Execute callback
	}

	// Log queue closure
	x_log.Debug().Msg("async callback queue closed")
}

// push adds a callback to the queue.
func (ac *asyncCallbacksHandler) push(f func()) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if ac.closed { // Warn if attempting to push after close
		x_log.Warn().Msg("async push attempted after close")
		return
	}

	ac.cbQueue <- f // Add callback to queue
}

// close shuts down the handler and marks it as closed.
func (ac *asyncCallbacksHandler) close() {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if ac.closed { // Log if close is called multiple times
		x_log.Debug().Msg("async close called multiple times")
		return
	}

	close(ac.cbQueue) // Close the callback queue
	ac.closed = true

	// Log dispatcher close
	x_log.Debug().Msg("async callback dispatcher closed")
}
