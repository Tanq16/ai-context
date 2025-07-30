package utils

import (
	"sync"
	"time"
)

type Console struct {
	status    string
	message   string
	progress  string
	complete  bool
	startTime time.Time
	mutex     sync.RWMutex
	err       error
	doneCh    chan struct{}  // Channel to signal stopping the display
	displayWg sync.WaitGroup // WaitGroup for display goroutine shutdown
}

func NewConsole() *Console {
	return &Console{
		status:    "pending",
		startTime: time.Now(),
		doneCh:    make(chan struct{}),
	}
}
