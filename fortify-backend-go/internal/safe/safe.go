// Package safe contains panic containment for the scanner's goroutine fan-out.
// A single panicking probe (malformed URL, nil map, regex blowup) must fail
// that probe — never the whole server process.
package safe

import "log"

// Go runs fn in a new goroutine, logging instead of crashing on panic.
// For fire-and-forget background work (scheduler ticks, script fetches).
func Go(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[safe] recovered background panic: %v", r)
			}
		}()
		fn()
	}()
}

// Send runs fn and delivers its value, delivering zero on panic so receivers
// never block forever. For fan-in channel patterns (runner results).
func Send[T any](ch chan<- T, zero T, fn func() T) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[safe] recovered worker panic: %v", r)
			ch <- zero
		}
	}()
	ch <- fn()
}

// Recover logs instead of crashing on panic. Deferred at the top of a
// goroutine whose partial results are already collected elsewhere, e.g.:
//
//	go func(p string) {
//	    defer wg.Done()
//	    defer safe.Recover() // a panicking param is skipped, rest continue
//	    ...
//	}(param)
func Recover() {
	if r := recover(); r != nil {
		log.Printf("[safe] recovered worker panic: %v", r)
	}
}

// Do runs fn, logging instead of crashing on panic. For worker-pool job
// bodies where skipping the item is the right fallback.
func Do(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[safe] recovered worker panic: %v", r)
		}
	}()
	fn()
}
