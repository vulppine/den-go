package den

import (
	"den/routing"
	"net"
	"net/http"
)

type Server struct {
	httpServer *http.Server
	Router     *routing.Router
}

func NewServer(ip net.Addr) *Server {
	router := new(routing.Router)
	httpServer := &http.Server{
		Addr:    ip.String(),
		Handler: router,
	}
	serv := &Server{
		httpServer,
		router,
	}

	return serv
}
