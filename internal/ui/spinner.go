package ui

import (
	"fmt"
	"sync"
	"time"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Spinner shows an animated wait line on stderr. Call the returned stop func when done.
func Spinner(label string) func() {
	if Default.Quiet || !colorEnabled() {
		return func() {}
	}

	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		tick := time.NewTicker(80 * time.Millisecond)
		defer tick.Stop()
		i := 0
		for {
			select {
			case <-done:
				return
			case <-tick.C:
				frame := spinnerFrames[i%len(spinnerFrames)]
				i++
				fmt.Fprintf(errOut, "\r  %s %s", render(mutedStyle, frame), label)
			}
		}
	}()

	var once sync.Once
	return func() {
		once.Do(func() {
			close(done)
			wg.Wait()
			fmt.Fprint(errOut, "\r\033[K")
		})
	}
}

// ProgressSpinner returns update/done funcs for changing the spinner label mid-operation.
func ProgressSpinner(initial string) (update func(string), done func()) {
	stop := Spinner(initial)
	var once sync.Once
	finish := func() {
		once.Do(func() { stop() })
	}
	return func(label string) {
		stop()
		stop = Spinner(label)
	}, finish
}
