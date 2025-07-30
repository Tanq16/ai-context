package utils

import (
	"log/slog"
	"sync"
)

type MSG struct {
	errType bool
	text    string
	args    []any
}

type Console struct {
	msgCh    chan MSG
	doneCh   chan struct{}
	wg       sync.WaitGroup
	mutex    sync.RWMutex
	disabled bool
}

func NewConsole() *Console {
	return &Console{
		msgCh:    make(chan MSG, 100),
		doneCh:   make(chan struct{}),
		disabled: false,
	}
}

func (c *Console) Disable() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.disabled = true
}

func (c *Console) Start() {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case msg, ok := <-c.msgCh:
				if !ok {
					return
				}
				c.mutex.RLock()
				disabled := c.disabled
				c.mutex.RUnlock()
				if !disabled {
					if msg.errType {
						slog.Error(msg.text, msg.args...)
					} else {
						slog.Info(msg.text, msg.args...)
					}
				}
			case <-c.doneCh:
				return
			}
		}
	}()
}

func (c *Console) Stop() {
	close(c.doneCh)
	close(c.msgCh)
	c.wg.Wait()
}

func (c *Console) Log(message string, errType bool, args ...any) {
	select {
	case c.msgCh <- MSG{text: message, errType: errType, args: args}:
	default:
		// Channel is full, drop message
	}
}
