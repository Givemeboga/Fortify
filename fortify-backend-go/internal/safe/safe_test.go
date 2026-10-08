package safe

import (
	"sync"
	"testing"
	"time"
)

func TestSendDeliversZeroOnPanic(t *testing.T) {
	ch := make(chan string, 1)
	go Send(ch, "fallback", func() string { panic("boom") })
	select {
	case got := <-ch:
		if got != "fallback" {
			t.Fatalf("got %q, want fallback", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("receiver blocked — panic escaped the worker")
	}
}

func TestGoSurvivesPanic(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	Go(func() {
		defer wg.Done()
		panic("boom")
	})
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("panic escaped Go()")
	}
}
