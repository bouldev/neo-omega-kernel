package can_close

import "sync"

type CanClose interface {
	// close
	Close()
	// wait until close and return error which cause close
	WaitClosed() <-chan error
	// is closed
	Closed() bool
	// error which cause close
	CloseError() error
}

type CanCloseWithError interface {
	CanClose
	CloseWithError(err error)
}

type Close struct {
	once       sync.Once
	close      chan struct{}
	closed     bool
	closeError error
	hook       func()
}

func NewClose(hook func()) CanCloseWithError {
	return &Close{
		hook:  hook,
		close: make(chan struct{}),
	}
}

func (c *Close) CloseWithError(err error) {
	c.once.Do(func() {
		c.closeError = err
		c.closed = true
		close(c.close)
		if c.hook != nil {
			c.hook()
		}
	})
}

func (c *Close) Close() {
	c.CloseWithError(nil)
}

func (c *Close) WaitClosed() <-chan error {
	ec := make(chan error)
	go func() {
		<-c.close
		ec <- c.closeError
	}()
	return ec
}

func (c *Close) Closed() bool {
	return c.closed
}

func (c *Close) CloseError() error {
	return c.closeError
}
