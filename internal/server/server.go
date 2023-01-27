package server

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func GetNewServer(ip string, port uint16, router *httprouter.Router) *http.Server {
	// Just returns default http server
	return &http.Server{
		Addr:    fmt.Sprintf("%s:%d", ip, port),
		Handler: router,
	}
}
