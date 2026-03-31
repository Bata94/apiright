package core

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type Spinner struct {
	mu       sync.Mutex
	message  string
	chars    []string
	interval time.Duration
	stopChan chan struct{}
	running  bool
}

var defaultSpinner = &Spinner{
	chars:    []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	interval: 100 * time.Millisecond,
}

func NewSpinner(message string) *Spinner {
	return &Spinner{
		message:  message,
		chars:    defaultSpinner.chars,
		interval: defaultSpinner.interval,
		stopChan: make(chan struct{}),
	}
}

func (s *Spinner) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	go func() {
		i := 0
		for {
			select {
			case <-s.stopChan:
				return
			default:
				s.mu.Lock()
				msg := s.message
				s.mu.Unlock()
				fmt.Printf("\r%s %s", s.chars[i%len(s.chars)], msg)
				i++
				time.Sleep(s.interval)
			}
		}
	}()
}

func (s *Spinner) SetMessage(msg string) {
	s.mu.Lock()
	s.message = msg
	s.mu.Unlock()
}

func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopChan)
	s.stopChan = make(chan struct{})
	clearLine()
}

func (s *Spinner) Success(msg string) {
	s.Stop()
	fmt.Printf("\r%s %s\n", Greenf("✓"), msg)
}

func (s *Spinner) Error(msg string) {
	s.Stop()
	fmt.Printf("\r%s %s\n", Redf("✗"), msg)
}

func (s *Spinner) Info(msg string) {
	s.Stop()
	fmt.Printf("\r%s %s\n", Cyanf("●"), msg)
}

func clearLine() {
	fmt.Print("\r" + strings.Repeat(" ", 80) + "\r")
}

func WithSpinner(message string, fn func() error) error {
	s := NewSpinner(message)
	s.Start()
	err := fn()
	if err != nil {
		s.Error(err.Error())
	} else {
		s.Success(message + " complete")
	}
	return err
}
