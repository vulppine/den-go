package routing

import (
	"time"
)

type routingContext struct {
	stage     routeStage // don't necessarily like this one
	stageChan chan routeStage
	info      *ResponseInfo
	data      *ResponseData
	err       error
	done      chan struct{}
}

func newRoutingContext() *routingContext {
	ctx := new(routingContext)
	ctx.stageChan = make(chan routeStage, 1)

	return ctx
}

func (c *routingContext) upgradeStage() {
	if c.stage >= finish {
		return
	}

	c.stageChan <- c.stage + 1
}

func (c *routingContext) CloseWithError(err error) {
	if c.err != nil {
		return
	}

	c.err = err

	c.done <- struct{}{}

	close(c.done)
}

func (c *routingContext) Deadline() (deadline time.Time, ok bool) {
	//TODO implement me
	panic("implement me")
}

func (c *routingContext) Err() error {
	return c.err
}

func (c *routingContext) Value(key any) any {
	//TODO implement me
	panic("implement me")
}

func (c *routingContext) Done() <-chan struct{} {
	return c.done
}
