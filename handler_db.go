package den

import (
	"den/routing"
	"errors"
)

// intense levels of reflection fuckery,
// so what this does is store all handlers
// inside a singleton that allows some kind
// of configuration format to fetch a handler
// and then create the handler based on the
// format:
//
// it requires interface{} casting and is
// generally Not Very Nice To Deal With internally
// but generally should be handled with care
// (and a lot of err/ok checks)
//
// the map[interface{}]interface{} is passed into the handler
// and what comes out should be... a handler
// spooky!

// DO NOT MODIFY AFTER GO INIT FUNCTIONS
// todo: add a write lock to this that's permanently locked
// todo: make the databases more generic

// RoutingHandlers is a database that stores all the routing handlers
var RoutingHandlers HandlerDatabase

type HandlerDatabase struct {
	handlers map[string]HandlerInit
}

type HandlerInit func(map[interface{}]interface{}) routing.RouteHandler

func (db *HandlerDatabase) Add(name string, init HandlerInit) error {
	if _, ok := db.handlers[name]; ok {
		return errors.New(name + " was already stored in the handler database")
	}

	db.handlers[name] = init
	return nil
}
