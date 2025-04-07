package core

import (
	"sync"

	"github.com/rskv-p/mini/pkg/x_log"
)

// asyncCallbacksHandler manages async callback queue.
type asyncCallbacksHandler struct {
	cbQueue chan func() // Queue of callbacks
	logger  x_log.Logger

	mu     sync.Mutex // Mutex for synchronization
	closed bool       // Indicates if handler is closed
}

// run starts executing callbacks from the queue.
func (ac *asyncCallbacksHandler) run() {
	if ac.logger != nil {
		ac.logger.Debug("async callback loop started") // Log callback loop start
	}

	for fn := range ac.cbQueue { // Process each callback in queue
		if fn == nil {
			if ac.logger != nil {
				ac.logger.Warn("nil callback received") // Log if nil callback
			}
			continue
		}
		fn() // Execute callback
	}

	if ac.logger != nil {
		ac.logger.Debug("async callback queue closed") // Log queue closure
	}
}

// push adds a callback to the queue.
func (ac *asyncCallbacksHandler) push(f func()) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if ac.closed { // Warn if attempting to push after close
		if ac.logger != nil {
			ac.logger.Warn("async push attempted after close")
		}
		return
	}

	ac.cbQueue <- f // Add callback to queue
}

// close shuts down the handler and marks it as closed.
func (ac *asyncCallbacksHandler) close() {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if ac.closed { // Log if close is called multiple times
		if ac.logger != nil {
			ac.logger.Debug("async close called multiple times")
		}
		return
	}

	close(ac.cbQueue) // Close the callback queue
	ac.closed = true

	if ac.logger != nil {
		ac.logger.Debug("async callback dispatcher closed") // Log dispatcher close
	}
}
